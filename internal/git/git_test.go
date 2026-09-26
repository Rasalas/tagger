package git

import (
	"os"
	"os/exec"
	"path/filepath"
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

// initTempRepo creates a throwaway git repository and chdirs into it.
func initTempRepo(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_COUNT", "0")
	t.Chdir(t.TempDir())
	gitCommand(t, "init", "-q", "-b", "main")
	gitCommand(t, "config", "user.email", "test@example.com")
	gitCommand(t, "config", "user.name", "Test")
	gitCommand(t, "config", "commit.gpgsign", "false")
	gitCommand(t, "config", "tag.gpgsign", "false")
	gitCommand(t, "commit", "--allow-empty", "--no-gpg-sign", "-m", "init")
}

func gitCommand(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func bareRemote(t *testing.T) string {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "remote.git")
	gitCommand(t, "init", "-q", "--bare", remote)
	gitCommand(t, "remote", "add", "origin", remote)
	return remote
}

func TestExecGitPushTagDoesNotFollowOtherAnnotatedTags(t *testing.T) {
	initTempRepo(t)
	remote := bareRemote(t)
	gitCommand(t, "config", "push.followTags", "true")
	gitCommand(t, "tag", "-a", "private-draft", "-m", "private")
	gitCommand(t, "tag", "-a", "v1.0.0", "-m", "release")

	if err := (ExecGit{}).PushTag("v1.0.0"); err != nil {
		t.Fatalf("PushTag: %v", err)
	}
	if got := gitCommand(t, "--git-dir", remote, "tag", "--list"); got != "v1.0.0" {
		t.Errorf("remote tags = %q, want only v1.0.0", got)
	}
}

func TestExecGitPushTagWithSameNamedBranch(t *testing.T) {
	initTempRepo(t)
	remote := bareRemote(t)
	gitCommand(t, "branch", "v1.0.0")
	gitCommand(t, "tag", "v1.0.0")

	if err := (ExecGit{}).PushTag("v1.0.0"); err != nil {
		t.Fatalf("PushTag with same-named branch: %v", err)
	}
	if got := gitCommand(t, "--git-dir", remote, "tag", "--list"); got != "v1.0.0" {
		t.Errorf("remote tags = %q, want v1.0.0", got)
	}
	if got := gitCommand(t, "--git-dir", remote, "for-each-ref", "--format=%(refname)", "refs/heads"); got != "" {
		t.Errorf("remote branches = %q, want none", got)
	}
}

func TestExecGitLatestTagSkipsUnreachableTags(t *testing.T) {
	initTempRepo(t)
	gitCommand(t, "tag", "v1.0.0")
	gitCommand(t, "switch", "-q", "-c", "other")
	gitCommand(t, "commit", "--allow-empty", "--no-gpg-sign", "-m", "feat: other")
	gitCommand(t, "tag", "v2.0.0")
	gitCommand(t, "switch", "-q", "main")
	gitCommand(t, "commit", "--allow-empty", "--no-gpg-sign", "-m", "fix: main")

	got, err := (ExecGit{}).LatestTag("v")
	if err != nil {
		t.Fatalf("LatestTag: %v", err)
	}
	if got != "v1.0.0" {
		t.Errorf("LatestTag = %q, want v1.0.0", got)
	}
}

func TestExecGitLatestTagWithColumnsEnabled(t *testing.T) {
	initTempRepo(t)
	gitCommand(t, "config", "column.tag", "always")
	gitCommand(t, "tag", "v1.0.0")
	gitCommand(t, "tag", "v1.1.0")

	got, err := (ExecGit{}).LatestTag("v")
	if err != nil {
		t.Fatalf("LatestTag: %v", err)
	}
	if got != "v1.1.0" {
		t.Errorf("LatestTag = %q, want v1.1.0", got)
	}
}

func TestExecGitLatestTagPreservesNumericBuildOrdering(t *testing.T) {
	initTempRepo(t)
	for _, tag := range []string{"v1.2.0-9", "v1.2.0-10"} {
		gitCommand(t, "tag", tag)
	}

	got, err := (ExecGit{}).LatestTag("v")
	if err != nil {
		t.Fatalf("LatestTag: %v", err)
	}
	if got != "v1.2.0-10" {
		t.Errorf("LatestTag = %q, want v1.2.0-10", got)
	}
}

func TestExecGitLatestTagSkipsUnparsableTags(t *testing.T) {
	initTempRepo(t)

	for _, tag := range []string{"v0.9.0", "not-a-version", "v-next", "v1.0.0"} {
		cmd := exec.Command("git", "tag", tag)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("tag %s: %v: %s", tag, err, out)
		}
	}

	tag, err := ExecGit{}.LatestTag("v")
	if err != nil {
		t.Fatalf("LatestTag error: %v", err)
	}
	if tag != "v1.0.0" {
		t.Errorf("LatestTag = %q, want %q (only valid semver tag)", tag, "v1.0.0")
	}
}

func TestExecGitLatestTagEmptyRepo(t *testing.T) {
	initTempRepo(t)

	tag, err := ExecGit{}.LatestTag("v")
	if err != nil {
		t.Fatalf("LatestTag error: %v", err)
	}
	if tag != "" {
		t.Errorf("LatestTag = %q, want empty for repo without tags", tag)
	}
}
