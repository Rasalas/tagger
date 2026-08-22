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
	RunE:  runBump,
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"suggest"},
	Short:   "Suggest the next semver tag without creating it",
	Long:    "Analyzes commits since the last tag using Conventional Commits and suggests the next semantic version tag.",
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

func runStatus(cmd *cobra.Command, args []string) error {
	return runBumpPlan(cmd, true, "Suggestion only — no changes made")
}

func runBump(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return runBumpPlan(cmd, dryRun, "Dry run — no changes made")
}

func runBumpPlan(cmd *cobra.Command, dryRun bool, dryRunMessage string) error {
	noPush := false
	yes := true
	if cmd.Flags().Lookup("no-push") != nil {
		noPush, _ = cmd.Flags().GetBool("no-push")
	}
	if cmd.Flags().Lookup("yes") != nil {
		yes, _ = cmd.Flags().GetBool("yes")
	}
	forceMajor, _ := cmd.Flags().GetBool("major")
	forceMinor, _ := cmd.Flags().GetBool("minor")
	forcePatch, _ := cmd.Flags().GetBool("patch")
	prefix, _ := cmd.Flags().GetString("prefix")

	// Validate: at most one force flag
	forceCount := 0
	if forceMajor {
		forceCount++
	}
	if forceMinor {
		forceCount++
	}
	if forcePatch {
		forceCount++
	}
	if forceCount > 1 {
		return configErr(fmt.Errorf("only one of --major, --minor, --patch can be specified"))
	}

	// Get latest tag
	latestTag, err := git.LatestTag(prefix)
	if err != nil {
		return fmt.Errorf("failed to get latest tag: %w", err)
	}
	hasLatestTag := latestTag != ""

	var current semver.Version
	if !hasLatestTag {
		current = semver.Version{}
		latestTag = current.Format(prefix)
		term.Info(fmt.Sprintf("No tags found, starting from %s", latestTag))
	} else {
		current, err = semver.ParseWithPrefix(latestTag, prefix)
		if err != nil {
			return configErr(fmt.Errorf("failed to parse tag %q: %w", latestTag, err))
		}
	}

	// Get commits since tag
	var commitTag string
	if hasLatestTag {
		commitTag = latestTag
	}
	messages, err := git.CommitsSince(commitTag)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}
	if len(messages) == 0 {
		term.Warn("No commits since " + latestTag)
		return nil
	}

	// Determine bump level
	var level semver.BumpLevel
	var commits []commit.Commit

	switch {
	case forceMajor:
		level = semver.Major
		_, commits = commit.Analyze(messages)
	case forceMinor:
		level = semver.Minor
		_, commits = commit.Analyze(messages)
	case forcePatch:
		level = semver.Patch
		_, commits = commit.Analyze(messages)
	default:
		level, commits = commit.Analyze(messages)
	}

	if level == semver.None {
		term.Warn("No conventional commits found, cannot determine bump")
		term.Info("Use --major, --minor, or --patch to force a bump")
		return nil
	}

	next := current.Bump(level)
	nextTag := next.Format(prefix)

	// Summary
	fmt.Fprintln(term.W)
	term.Header("tagger")
	fmt.Fprintf(term.W, "  %sCurrent%s  %s\n", term.Muted, term.Reset, latestTag)
	fmt.Fprintf(term.W, "  %sCommits%s  %d\n", term.Muted, term.Reset, len(commits))
	fmt.Fprintf(term.W, "  %sBump%s     %s\n", term.Muted, term.Reset, level)
	fmt.Fprintf(term.W, "  %sNext%s     %s%s%s\n", term.Muted, term.Reset, term.Primary, nextTag, term.Reset)
	fmt.Fprintln(term.W)

	// Show commits
	for _, c := range commits {
		cl := commit.Classify(c)
		switch {
		case c.Breaking:
			term.Fail(formatCommit(c))
		case cl == semver.Minor:
			term.Pass(formatCommit(c))
		case cl == semver.Patch:
			term.Info(formatCommit(c))
		default:
			fmt.Fprintf(term.W, "  %s·%s %s\n", term.Muted, term.Reset, formatCommit(c))
		}
	}
	fmt.Fprintln(term.W)

	if dryRun {
		term.Info(dryRunMessage)
		return nil
	}

	// Confirmation
	if !yes {
		if !confirmIsTTY() {
			return fmt.Errorf("non-interactive session: use --yes to create %s without prompting", nextTag)
		}
		if !confirm(nextTag) {
			return nil
		}
	}

	// Create tag
	if err := git.CreateTag(nextTag, "Release "+nextTag); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	term.Pass("Created " + nextTag)

	// Push (default: always push, skip with --no-push)
	if !noPush {
		if err := git.PushTag(nextTag); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}
		term.Pass("Pushed " + nextTag)
	}

	fmt.Fprintln(term.W)
	return nil
}

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

func formatCommit(c commit.Commit) string {
	if c.Type == "" {
		return c.Summary
	}
	if c.Scope != "" {
		return fmt.Sprintf("%s(%s): %s", c.Type, c.Scope, c.Summary)
	}
	return fmt.Sprintf("%s: %s", c.Type, c.Summary)
}
