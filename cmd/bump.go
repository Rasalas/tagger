package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rasalas/tagger/internal/commit"
	"github.com/rasalas/tagger/internal/git"
	"github.com/rasalas/tagger/internal/semver"
	"github.com/rasalas/tagger/internal/term"
	"github.com/spf13/cobra"
)

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Determine and create the next semver tag",
	Long:  "Analyzes commits since the last tag using Conventional Commits and creates the next semantic version tag.",
	RunE:  runBump,
}

func init() {
	bumpCmd.Flags().BoolP("push", "p", false, "Push tag after creation")
	bumpCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	bumpCmd.Flags().Bool("dry-run", false, "Show what would happen without making changes")
	bumpCmd.Flags().Bool("major", false, "Force a major bump")
	bumpCmd.Flags().Bool("minor", false, "Force a minor bump")
	bumpCmd.Flags().Bool("patch", false, "Force a patch bump")
	bumpCmd.Flags().String("prefix", "v", "Tag prefix")
	rootCmd.AddCommand(bumpCmd)
}

func runBump(cmd *cobra.Command, args []string) error {
	push, _ := cmd.Flags().GetBool("push")
	yes, _ := cmd.Flags().GetBool("yes")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
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
		return fmt.Errorf("only one of --major, --minor, --patch can be specified")
	}

	// Get latest tag
	latestTag, err := git.LatestTag(prefix)
	if err != nil {
		return fmt.Errorf("failed to get latest tag: %w", err)
	}

	var current semver.Version
	if latestTag == "" {
		current = semver.Version{}
		latestTag = current.Format(prefix)
		term.Info(fmt.Sprintf("No tags found, starting from %s", latestTag))
	} else {
		current, err = semver.Parse(latestTag)
		if err != nil {
			return fmt.Errorf("failed to parse tag %q: %w", latestTag, err)
		}
	}

	// Get commits since tag
	var commitTag string
	if current == (semver.Version{}) {
		commitTag = "" // no previous tag, get all commits
	} else {
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
		term.Info("Dry run — no changes made")
		return nil
	}

	// Confirmation
	if !yes {
		fmt.Fprintf(term.W, "  Create %s%s%s? [Y/n] ", term.Primary, nextTag, term.Reset)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			term.Warn("Aborted")
			return nil
		}
	}

	// Create tag
	if err := git.CreateTag(nextTag, "Release "+nextTag); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	term.Pass("Created " + nextTag)

	// Push
	if push {
		if err := git.PushTag(nextTag); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}
		term.Pass("Pushed " + nextTag)
	}

	fmt.Fprintln(term.W)
	return nil
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
