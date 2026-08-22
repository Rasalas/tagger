package cmd

import (
	"bytes"
	"errors"
	"os"
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
	origConfirmIn := confirmIn
	origConfirmIsTTY := confirmIsTTY
	buf := &bytes.Buffer{}
	git.Default = mock
	term.W = buf
	confirmIn = os.Stdin
	confirmIsTTY = func() bool { return true }
	return buf, func() {
		git.Default = origGit
		term.W = origW
		confirmIn = origConfirmIn
		confirmIsTTY = origConfirmIsTTY
		// Reset cobra flags to defaults between tests
		bumpCmd.Flags().Set("no-push", "false")
		bumpCmd.Flags().Set("yes", "false")
		bumpCmd.Flags().Set("dry-run", "false")
		bumpCmd.Flags().Set("major", "false")
		bumpCmd.Flags().Set("minor", "false")
		bumpCmd.Flags().Set("patch", "false")
		bumpCmd.Flags().Set("prefix", "v")
		statusCmd.Flags().Set("major", "false")
		statusCmd.Flags().Set("minor", "false")
		statusCmd.Flags().Set("patch", "false")
		statusCmd.Flags().Set("prefix", "v")
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
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Errorf("force flag conflict should be a ConfigError, got %T", err)
	}
}

func TestBumpInvalidTagIsConfigError(t *testing.T) {
	mock := &mockGit{
		latestTag: "not-a-version",
		commits:   []string{"feat: something"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--yes"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unparsable tag")
	}
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Errorf("unparsable tag should be a ConfigError, got %T: %v", err, err)
	}
}

func TestBumpGitFailureIsRuntimeError(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: add feature"},
		createErr: errors.New("exit status 128: permission denied"),
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--yes"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when CreateTag fails")
	}
	var cfgErr *ConfigError
	if errors.As(err, &cfgErr) {
		t.Errorf("git runtime failure must not be a ConfigError, got: %v", err)
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

func TestBumpIncrementsBuildTag(t *testing.T) {
	mock := &mockGit{
		latestTag: "v3.4.0-58",
		commits:   []string{"fix: repair export"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v3.4.1-59" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v3.4.1-59")
	}
	if mock.pushedTag != "v3.4.1-59" {
		t.Errorf("pushed tag = %q, want %q", mock.pushedTag, "v3.4.1-59")
	}
}

func TestBumpParsesConfiguredPrefix(t *testing.T) {
	mock := &mockGit{
		latestTag: "release-3.4.0-58",
		commits:   []string{"fix: repair export"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"bump", "--prefix", "release-", "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "release-3.4.1-59" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "release-3.4.1-59")
	}
}

func TestStatusSuggestsWithoutCreatingTag(t *testing.T) {
	mock := &mockGit{
		latestTag: "v3.4.0-58",
		commits:   []string{"feat: add report filters"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "v3.5.0-59") {
		t.Errorf("expected next tag v3.5.0-59 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Suggestion only") {
		t.Errorf("expected suggestion message in output, got:\n%s", out)
	}
	if mock.createdTag != "" {
		t.Errorf("status should not create a tag, created %q", mock.createdTag)
	}
	if mock.pushedTag != "" {
		t.Errorf("status should not push a tag, pushed %q", mock.pushedTag)
	}
}

func TestSuggestAlias(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.2.3",
		commits:   []string{"fix: repair export"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()

	rootCmd.SetArgs([]string{"suggest"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "v1.2.4") {
		t.Errorf("expected next tag v1.2.4 in output, got:\n%s", out)
	}
	if mock.createdTag != "" {
		t.Errorf("suggest should not create a tag, created %q", mock.createdTag)
	}
}

func TestBumpRefusesNonInteractiveWithoutYes(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: add feature"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()
	confirmIsTTY = func() bool { return false }

	rootCmd.SetArgs([]string{"bump"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error in non-interactive session without --yes")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("error should hint at --yes, got: %v", err)
	}
	if mock.createdTag != "" {
		t.Errorf("no tag should be created without confirmation, created %q", mock.createdTag)
	}
	_ = buf
}

func TestBumpPromptEOFAborts(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: add feature"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()
	confirmIn = strings.NewReader("")

	rootCmd.SetArgs([]string{"bump"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Aborted") {
		t.Errorf("expected abort message on EOF, got:\n%s", out)
	}
	if mock.createdTag != "" {
		t.Errorf("EOF must never create a tag, created %q", mock.createdTag)
	}
}

func TestBumpPromptDeclineAborts(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: add feature"},
	}
	buf, cleanup := setupTest(mock)
	defer cleanup()
	confirmIn = strings.NewReader("n\n")

	rootCmd.SetArgs([]string{"bump"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Aborted") {
		t.Errorf("expected abort message on decline, got:\n%s", out)
	}
	if mock.createdTag != "" {
		t.Errorf("no tag should be created after decline, created %q", mock.createdTag)
	}
}

func TestBumpPromptAcceptCreatesTag(t *testing.T) {
	mock := &mockGit{
		latestTag: "v1.0.0",
		commits:   []string{"feat: add feature"},
	}
	_, cleanup := setupTest(mock)
	defer cleanup()
	confirmIn = strings.NewReader("y\n")

	rootCmd.SetArgs([]string{"bump"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.createdTag != "v1.1.0" {
		t.Errorf("created tag = %q, want %q", mock.createdTag, "v1.1.0")
	}
}
