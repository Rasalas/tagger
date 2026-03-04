package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rasalas/tagger/internal/git"
	"github.com/rasalas/tagger/internal/term"
)

type mockGit struct {
	latestTag  string
	latestErr  error
	commits    []string
	commitsErr error
	createErr  error
	pushErr    error

	createdTag string
	createdMsg string
	pushedTag  string
}

func (m *mockGit) LatestTag(prefix string) (string, error) {
	return m.latestTag, m.latestErr
}

func (m *mockGit) CommitsSince(tag string) ([]string, error) {
	return m.commits, m.commitsErr
}

func (m *mockGit) CreateTag(tag, message string) error {
	m.createdTag = tag
	m.createdMsg = message
	return m.createErr
}

func (m *mockGit) PushTag(tag string) error {
	m.pushedTag = tag
	return m.pushErr
}

func setupTest(mock *mockGit) (*bytes.Buffer, func()) {
	origGit := git.Default
	origW := term.W
	buf := &bytes.Buffer{}
	git.Default = mock
	term.W = buf
	return buf, func() {
		git.Default = origGit
		term.W = origW
		// Reset cobra flags to defaults between tests
		bumpCmd.Flags().Set("no-push", "false")
		bumpCmd.Flags().Set("yes", "false")
		bumpCmd.Flags().Set("dry-run", "false")
		bumpCmd.Flags().Set("major", "false")
		bumpCmd.Flags().Set("minor", "false")
		bumpCmd.Flags().Set("patch", "false")
		bumpCmd.Flags().Set("prefix", "v")
	}
}

func TestBumpDryRun(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: add feature"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--dry-run", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "v1.1.0") {
		t.Errorf("expected next tag v1.1.0 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run message in output, got:\n%s", out)
	}
	if mock.createdTag != "" {
		t.Error("tag should not have been created in dry run")
	}
}

func TestBumpNoCommits(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   nil,
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No commits") {
		t.Errorf("expected no commits warning, got:\n%s", out)
	}
}

func TestBumpNoConventionalCommits(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"update readme", "misc changes"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No conventional commits") {
		t.Errorf("expected no conventional commits warning, got:\n%s", out)
	}
}

func TestBumpForcePatch(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"update readme"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--patch", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v1.0.1" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v1.0.1")
	}
	out := buf.String()
	if !strings.Contains(out, "Created v1.0.1") {
		t.Errorf("expected created message, got:\n%s", out)
	}
}

func TestBumpForceMinor(t *testing.T) {
	mock := &mockGit{
		latestTag: "v2.3.1",
		commits:   []string{"fix: something"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--minor", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v2.4.0" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v2.4.0")
	}
	_ = buf
}

func TestBumpForceMajor(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.2.3",
		commits:   []string{"fix: something"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--major", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v2.0.0" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v2.0.0")
	}
}

func TestBumpMultipleForceFlags(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: something"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--major", "--minor", "--yes"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for multiple force flags")
	}
	if !strings.Contains(err.Error(), "only one") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBumpNoTag(t *testing.T) {
	mock := &mockGit{
		latestTag: "",
		commits:   []string{"feat: initial feature"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v0.1.0" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v0.1.0")
	}
}

func TestBumpPushesByDefault(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: add feature"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v1.1.0" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v1.1.0")
	}
	if mock.pushedTag != "v1.1.0" {
		t.Errorf("pushed tag = %q, want %q", mock.pushedTag, "v1.1.0")
	}
	out := buf.String()
	if !strings.Contains(out, "Pushed v1.1.0") {
		t.Errorf("expected pushed message, got:\n%s", out)
	}
}

func TestBumpNoPush(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: add feature"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--no-push", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v1.1.0" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v1.1.0")
	}
	if mock.pushedTag != "" {
		t.Errorf("should not have pushed, but pushed %q", mock.pushedTag)
	}
}

func TestBumpBreakingCommit(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.2.3",
		commits:   []string{"feat!: breaking change"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v2.0.0" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v2.0.0")
	}
}
