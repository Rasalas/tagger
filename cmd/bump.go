package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rasalas/tagger/internal/commit"
	"github.com/rasalas/tagger/internal/git"
	"github.com/rasalas/tagger/internal/semver"
	"github.com/rasalas/tagger/internal/term"
	"github.com/spf13/cobra"
)

// confirmIn is the source for confirmation input, swappable for testing.
var confirmIn io.Reader = os.Stdin

// confirmIsTTY reports whether the confirmation input is an interactive
// terminal, swappable for testing.
var confirmIsTTY = func() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Determine and create the next semver tag",
	Long:  "Analyzes commits since the last tag using Conventional Commits and creates the next semantic version tag.",
	Args:  noArgs,
	RunE:  runBump,
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"suggest"},
	Short:   "Suggest the next semver tag without creating it",
	Long:    "Analyzes commits since the last tag using Conventional Commits and suggests the next semantic version tag.",
	Args:    noArgs,
	RunE:    runStatus,
}

func init() {
	bumpCmd.Flags().Bool("no-push", false, "Skip pushing the tag to origin")
	bumpCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	bumpCmd.Flags().Bool("dry-run", false, "Show what would happen without making changes")
	addAnalysisFlags(bumpCmd)
	rootCmd.AddCommand(bumpCmd)

	addAnalysisFlags(statusCmd)
	rootCmd.AddCommand(statusCmd)
}

func addAnalysisFlags(c *cobra.Command) {
	c.Flags().Bool("major", false, "Force a major bump")
	c.Flags().Bool("minor", false, "Force a minor bump")
	c.Flags().Bool("patch", false, "Force a patch bump")
	c.Flags().String("prefix", "v", "Tag prefix")
}

// bumpOptions holds everything needed to build and execute a bump plan.
type bumpOptions struct {
	prefix    string
	force     semver.BumpLevel // None means determine from commits
	noPush    bool
	assumeYes bool
}

// bumpPlan is the result of analyzing the repository.
type bumpPlan struct {
	currentTag string
	hasLatest  bool
	level      semver.BumpLevel
	commits    []commit.Commit
	current    semver.Version
	next       semver.Version
	nextTag    string
}

// loadAnalysisOptions reads the flags shared by bump and status.
func loadAnalysisOptions(cmd *cobra.Command) (bumpOptions, error) {
	forceMajor, _ := cmd.Flags().GetBool("major")
	forceMinor, _ := cmd.Flags().GetBool("minor")
	forcePatch, _ := cmd.Flags().GetBool("patch")

	forceCount := 0
	for _, f := range []bool{forceMajor, forceMinor, forcePatch} {
		if f {
			forceCount++
		}
	}
	if forceCount > 1 {
		return bumpOptions{}, configErr(fmt.Errorf("only one of --major, --minor, --patch can be specified"))
	}

	prefix, _ := cmd.Flags().GetString("prefix")
	opts := bumpOptions{prefix: prefix}
	switch {
	case forceMajor:
		opts.force = semver.Major
	case forceMinor:
		opts.force = semver.Minor
	case forcePatch:
		opts.force = semver.Patch
	}
	return opts, nil
}

// loadBumpOptions reads the analysis flags plus the lifecycle flags that only
// exist on the bump command.
func loadBumpOptions(cmd *cobra.Command) (bumpOptions, error) {
	opts, err := loadAnalysisOptions(cmd)
	if err != nil {
		return opts, err
	}
	opts.noPush, _ = cmd.Flags().GetBool("no-push")
	opts.assumeYes, _ = cmd.Flags().GetBool("yes")
	return opts, nil
}

// buildPlan gathers tags and commits and computes the next tag. It returns a
// nil plan when there is nothing to do; the reason is reported via term output.
func buildPlan(opts bumpOptions) (*bumpPlan, error) {
	tag, err := git.LatestTag(opts.prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag: %w", err)
	}

	p := &bumpPlan{currentTag: tag, hasLatest: tag != ""}
	if !p.hasLatest {
		p.current = semver.Version{}
		p.currentTag = p.current.Format(opts.prefix)
		term.Info(fmt.Sprintf("No tags found, starting from %s", p.currentTag))
	} else {
		v, err := semver.ParseWithPrefix(tag, opts.prefix)
		if err != nil {
			return nil, configErr(fmt.Errorf("failed to parse tag %q: %w", tag, err))
		}
		p.current = v
	}

	commitTag := ""
	if p.hasLatest {
		commitTag = tag
	}
	messages, err := git.CommitsSince(commitTag)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}
	if len(messages) == 0 {
		term.Warn("No commits since " + p.currentTag)
		return nil, nil
	}

	level, commits := commit.Analyze(messages)
	switch opts.force {
	case semver.Major:
		level = semver.Major
	case semver.Minor:
		level = semver.Minor
	case semver.Patch:
		level = semver.Patch
	}
	p.level = level
	p.commits = commits

	if level == semver.None {
		term.Warn("No conventional commits found, cannot determine bump")
		term.Info("Use --major, --minor, or --patch to force a bump")
		return nil, nil
	}

	p.next = p.current.Bump(level)
	p.nextTag = p.next.Format(opts.prefix)
	return p, nil
}

// printPlan renders the summary and the analyzed commits.
func printPlan(w io.Writer, p *bumpPlan) {
	fmt.Fprintln(w)
	term.Header("tagger")
	fmt.Fprintf(w, "  %sCurrent%s  %s\n", term.Muted, term.Reset, p.currentTag)
	fmt.Fprintf(w, "  %sCommits%s  %d\n", term.Muted, term.Reset, len(p.commits))
	fmt.Fprintf(w, "  %sBump%s     %s\n", term.Muted, term.Reset, p.level)
	fmt.Fprintf(w, "  %sNext%s     %s%s%s\n", term.Muted, term.Reset, term.Primary, p.nextTag, term.Reset)
	fmt.Fprintln(w)

	for _, c := range p.commits {
		cl := commit.Classify(c)
		switch {
		case c.Breaking:
			term.Fail(formatCommit(c))
		case cl == semver.Minor:
			term.Pass(formatCommit(c))
		case cl == semver.Patch:
			term.Info(formatCommit(c))
		default:
			fmt.Fprintf(w, "  %s·%s %s\n", term.Muted, term.Reset, formatCommit(c))
		}
	}
	fmt.Fprintln(w)
}

// executePlan creates the tag (after confirmation unless assumeYes) and pushes
// it unless noPush is set.
func executePlan(opts bumpOptions, p *bumpPlan) error {
	if !opts.assumeYes {
		if !confirmIsTTY() {
			return fmt.Errorf("non-interactive session: use --yes to create %s without prompting", p.nextTag)
		}
		if !confirm(p.nextTag) {
			return nil
		}
	}

	if err := git.CreateTag(p.nextTag, "Release "+p.nextTag); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	term.Pass("Created " + p.nextTag)

	// Push (default: always push, skip with --no-push)
	if !opts.noPush {
		if err := git.PushTag(p.nextTag); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}
		term.Pass("Pushed " + p.nextTag)
	}

	fmt.Fprintln(term.W)
	return nil
}

// confirm asks the user to approve the given tag. EOF or any answer other
// than y/yes aborts — silence never consents.
func confirm(tag string) bool {
	fmt.Fprintf(term.W, "  Create %s%s%s? [Y/n] ", term.Primary, tag, term.Reset)
	answer, err := bufio.NewReader(confirmIn).ReadString('\n')
	if err != nil && strings.TrimSpace(answer) == "" {
		// EOF or read error without input — treat as abort, never as consent.
		fmt.Fprintln(term.W)
		term.Warn("Aborted")
		return false
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "" && answer != "y" && answer != "yes" {
		term.Warn("Aborted")
		return false
	}
	return true
}

func runStatus(cmd *cobra.Command, args []string) error {
	opts, err := loadAnalysisOptions(cmd)
	if err != nil {
		return err
	}
	p, err := buildPlan(opts)
	if err != nil || p == nil {
		return err
	}
	printPlan(term.W, p)
	term.Info("Suggestion only — no changes made")
	return nil
}

func runBump(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	opts, err := loadBumpOptions(cmd)
	if err != nil {
		return err
	}
	p, err := buildPlan(opts)
	if err != nil || p == nil {
		return err
	}
	printPlan(term.W, p)

	if dryRun {
		term.Info("Dry run — no changes made")
		return nil
	}
	return executePlan(opts, p)
}

func formatCommit(c commit.Commit) string {
	if c.Type == "" {
		return c.Summary
	}
	if c.Scope != "" {
		return fmt.Sprintf("%s(%s): %s", c.Type, c.Scope, c.Summary)
	}
	return fmt.Sprintf("%s: %s", c.Type, c.Summary)
}
