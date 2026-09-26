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
	initTempRepo(t)
	for _, tag := range []string{"v1.0.0", "v999.00.0", "v999.+0.0"} {
		cmd := exec.Command("git", "tag", tag)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git tag %q: %v: %s", tag, err, out)
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
