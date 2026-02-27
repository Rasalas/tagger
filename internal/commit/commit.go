package commit

import (
	"strings"

	"github.com/rasalas/tagger/internal/semver"
)

// Commit represents a parsed Conventional Commit.
type Commit struct {
	Type     string
	Scope    string
	Breaking bool
	Summary  string
	Body     string
}

// Parse parses a raw commit message into a Commit.
func Parse(message string) Commit {
	first, body, _ := strings.Cut(message, "\n")
	first = strings.TrimSpace(first)
	body = strings.TrimSpace(body)

	c := Commit{Body: body}

	// Check for BREAKING CHANGE in body
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "BREAKING CHANGE:") || strings.HasPrefix(line, "BREAKING-CHANGE:") {
			c.Breaking = true
			break
		}
	}

	// Split on first ':'
	prefix, summary, found := strings.Cut(first, ":")
	if !found {
		c.Summary = first
		return c
	}
	c.Summary = strings.TrimSpace(summary)

	// Check for '!' before the colon
	if strings.HasSuffix(prefix, "!") {
		c.Breaking = true
		prefix = prefix[:len(prefix)-1]
	}

	// Extract scope from "type(scope)"
	if open := strings.Index(prefix, "("); open != -1 {
		if close := strings.Index(prefix[open:], ")"); close != -1 {
			c.Scope = prefix[open+1 : open+close]
			c.Type = prefix[:open]
		} else {
			c.Type = prefix
		}
	} else {
		c.Type = prefix
	}

	c.Type = strings.TrimSpace(c.Type)
	return c
}

// Classify returns the bump level for a single commit.
func Classify(c Commit) semver.BumpLevel {
	if c.Breaking {
		return semver.Major
	}
	switch c.Type {
	case "feat":
		return semver.Minor
	case "fix", "refactor", "perf", "revert":
		return semver.Patch
	default:
		return semver.None
	}
}

// Analyze parses all messages and returns the highest bump level and parsed commits.
func Analyze(messages []string) (semver.BumpLevel, []Commit) {
	var highest semver.BumpLevel
	commits := make([]Commit, 0, len(messages))
	for _, msg := range messages {
		c := Parse(msg)
		commits = append(commits, c)
		if level := Classify(c); level > highest {
			highest = level
		}
	}
	return highest, commits
}
