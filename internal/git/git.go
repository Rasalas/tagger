package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Git defines the operations needed for tagging.
type Git interface {
	LatestTag(prefix string) (string, error)
	CommitsSince(tag string) ([]string, error)
	CreateTag(tag, message string) error
	PushTag(tag string) error
}

// Default is the active Git implementation, swappable for testing.
var Default Git = ExecGit{}

// ExecGit shells out to the real git CLI.
type ExecGit struct{}

func run(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("git %s: %w: %s", args[0], err, msg)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// LatestTag returns the most recent tag matching the prefix, or "" if none.
func (ExecGit) LatestTag(prefix string) (string, error) {
	out, err := run("tag", "--list", prefix+"*", "--sort=-version:refname")
	if err != nil {
		return "", err
	}
	if out == "" {
		return "", nil
	}
	first, _, _ := strings.Cut(out, "\n")
	return first, nil
}

// CommitsSince returns commit messages since the given tag (or all if tag is empty).
func (ExecGit) CommitsSince(tag string) ([]string, error) {
	var rangeSpec string
	if tag != "" {
		rangeSpec = tag + "..HEAD"
	} else {
		rangeSpec = "HEAD"
	}
	out, err := run("log", rangeSpec, "--format=%B%x00")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var messages []string
	for _, part := range strings.Split(out, "\x00") {
		msg := strings.TrimSpace(part)
		if msg != "" {
			messages = append(messages, msg)
		}
	}
	return messages, nil
}

// CreateTag creates an annotated tag.
func (ExecGit) CreateTag(tag, message string) error {
	_, err := run("tag", "-a", tag, "-m", message)
	return err
}

// PushTag pushes a single tag to origin.
func (ExecGit) PushTag(tag string) error {
	_, err := run("push", "origin", tag)
	return err
}

// Free functions delegate to Default.

func LatestTag(prefix string) (string, error)     { return Default.LatestTag(prefix) }
func CommitsSince(tag string) ([]string, error)    { return Default.CommitsSince(tag) }
func CreateTag(tag, message string) error          { return Default.CreateTag(tag, message) }
func PushTag(tag string) error                     { return Default.PushTag(tag) }
