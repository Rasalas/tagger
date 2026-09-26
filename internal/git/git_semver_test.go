package git

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestLatestTagIgnoresMalformedNumericTags(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_COUNT", "0")
	t.Chdir(t.TempDir())
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.name", "Test"},
		{"config", "user.email", "test@example.com"},
		{"commit", "--allow-empty", "--no-gpg-sign", "-m", "init"},
		{"tag", "v1.0.0"},
		{"tag", "v999.00.0"},
		{"tag", "v999.+0.0"},
	} {
		cmd := exec.Command("git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	got, err := (ExecGit{}).LatestTag("v")
	if err != nil {
		t.Fatalf("LatestTag: %v", err)
	}
	if got != "v1.0.0" {
		t.Errorf("LatestTag = %q, want v1.0.0", strings.TrimSpace(got))
	}
}
