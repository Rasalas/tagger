package commit

import (
	"strings"
	"unicode"

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

	c := Commit{Summary: first, Body: body}
	if !parseGitLabMerge(first, &c) && !parseConventionalHeader(first, &c) {
		return c
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "BREAKING CHANGE:") || strings.HasPrefix(line, "BREAKING-CHANGE:") {
			c.Breaking = true
			break
		}
	}
	return c
}

func parseConventionalHeader(first string, c *Commit) bool {
	prefix, summary, found := strings.Cut(first, ":")
	if !found {
		return false
	}
	if !strings.HasPrefix(summary, " ") || strings.TrimSpace(summary) == "" {
		return false
	}

	// Check for '!' before the colon
	breaking := strings.HasSuffix(prefix, "!")
	if breaking {
		prefix = prefix[:len(prefix)-1]
	}

	// The entire prefix must be a type or type(scope), with no trailing text.
	typ, scope := prefix, ""
	if open := strings.Index(prefix, "("); open != -1 {
		if !strings.HasSuffix(prefix, ")") {
			return false
		}
		scope = prefix[open+1 : len(prefix)-1]
		typ = prefix[:open]
		if !validHeaderToken(scope) {
			return false
		}
	}
	if !validType(typ) {
		return false
	}

	// Only mutate the commit once the header is known to be valid.
	c.Summary = strings.TrimSpace(summary)
	c.Breaking = c.Breaking || breaking
	c.Scope = scope
	c.Type = strings.ToLower(typ)
	return true
}

func validHeaderToken(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsControl(r) || strings.ContainsRune("():!", r) {
			return false
		}
	}
	return true
}

func validType(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || i > 0 && (c >= '0' && c <= '9' || c == '-' || c == '_') {
			continue
		}
		return false
	}
	return true
}

func parseGitLabMerge(first string, c *Commit) bool {
	const prefix = "Merge branch '"
	if !strings.HasPrefix(first, prefix) {
		return false
	}

	rest := strings.TrimPrefix(first, prefix)
	branch, suffix, found := strings.Cut(rest, "'")
	if !found || !strings.HasPrefix(strings.TrimSpace(suffix), "into '") {
		return false
	}

	commitPrefix, summary, found := strings.Cut(strings.TrimSpace(branch), "/")
	if !found || strings.TrimSpace(commitPrefix) == "" || strings.TrimSpace(summary) == "" {
		return false
	}

	return parseConventionalHeader(commitPrefix+": "+summary, c)
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
