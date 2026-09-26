package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestBumpOverflowDoesNotCreateOrPushTag(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name, tag, message, component string
	}{
		{"major", fmt.Sprintf("v%d.2.3", maxInt), "feat!: breaking", "major"},
		{"minor", fmt.Sprintf("v1.%d.3", maxInt), "feat: new", "minor"},
		{"patch", fmt.Sprintf("v1.2.%d", maxInt), "fix: repair", "patch"},
		{"build", fmt.Sprintf("v1.2.3-%d", maxInt), "fix: repair", "build"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockGit{latestTag: tc.tag, commits: []string{tc.message}}
			_, cleanup := setupTest(mock)
			defer cleanup()
			rootCmd.SetArgs([]string{"bump", "--yes"})
			err := rootCmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.component) || !strings.Contains(err.Error(), "overflow") {
				t.Errorf("error = %v, want actionable %s overflow error", err, tc.component)
			}
			if mock.createdTag != "" || mock.pushedTag != "" {
				t.Errorf("created %q and pushed %q after overflow", mock.createdTag, mock.pushedTag)
			}
		})
	}
}
