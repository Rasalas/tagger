package git

import (
	"strings"
	"testing"
)

type mockGit struct {
	latestTag  string
	latestErr  error
	commits    []string
	commitsErr error
	createErr  error
	pushErr    error

	createCalledTag string
	createCalledMsg string
	pushCalledTag   string
	latestPrefix    string
	commitsSinceTag string
}

func (m *mockGit) LatestTag(prefix string) (string, error) {
	m.latestPrefix = prefix
	return m.latestTag, m.latestErr
}

func (m *mockGit) CommitsSince(tag string) ([]string, error) {
	m.commitsSinceTag = tag
	return m.commits, m.commitsErr
}

func (m *mockGit) CreateTag(tag, message string) error {
	m.createCalledTag = tag
	m.createCalledMsg = message
	return m.createErr
}

func (m *mockGit) PushTag(tag string) error {
	m.pushCalledTag = tag
	return m.pushErr
}

func TestFreeFunctionsDelegateToDefault(t *testing.T) {
	original := Default
	defer func() { Default = original }()

	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: something"},
	}
	Default = mock

	tag, err := LatestTag("v")
	if err != nil {
		t.Fatalf("LatestTag error: %v", err)
	}
	if tag != "v1.0.0" {
		t.Errorf("LatestTag = %q, want %q", tag, "v1.0.0")
	}
	if mock.latestPrefix != "v" {
		t.Errorf("prefix = %q, want %q", mock.latestPrefix, "v")
	}

	msgs, err := CommitsSince("v1.0.0")
	if err != nil {
		t.Fatalf("CommitsSince error: %v", err)
	}
	if len(msgs) != 1 || msgs[0] != "feat: something" {
		t.Errorf("CommitsSince = %v, want [feat: something]", msgs)
	}
	if mock.commitsSinceTag != "v1.0.0" {
		t.Errorf("commitsSinceTag = %q, want %q", mock.commitsSinceTag, "v1.0.0")
	}

	if err := CreateTag("v1.1.0", "Release v1.1.0"); err != nil {
		t.Fatalf("CreateTag error: %v", err)
	}
	if mock.createCalledTag != "v1.1.0" {
		t.Errorf("createCalledTag = %q, want %q", mock.createCalledTag, "v1.1.0")
	}
	if mock.createCalledMsg != "Release v1.1.0" {
		t.Errorf("createCalledMsg = %q, want %q", mock.createCalledMsg, "Release v1.1.0")
	}

	if err := PushTag("v1.1.0"); err != nil {
		t.Fatalf("PushTag error: %v", err)
	}
	if mock.pushCalledTag != "v1.1.0" {
		t.Errorf("pushCalledTag = %q, want %q", mock.pushCalledTag, "v1.1.0")
	}
}

func TestRunIncludesGitStderrOnError(t *testing.T) {
	_, err := run("rev-parse", "--verify", "definitely-not-a-real-ref")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
	msg := err.Error()
	if !strings.Contains(msg, "git rev-parse") {
		t.Errorf("error should name the failing subcommand, got: %v", msg)
	}
	if !strings.Contains(msg, "fatal:") {
		t.Errorf("error should include git stderr output, got: %v", msg)
	}
}
