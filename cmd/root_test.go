package cmd

import (
	"runtime/debug"
	"testing"
)

func TestVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{
			name: "module version",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v1.2.1"}},
			ok:   true,
			want: "v1.2.1",
		},
		{
			name: "local development build",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: version,
		},
		{
			name: "missing build info",
			info: nil,
			ok:   false,
			want: version,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := versionFromBuildInfo(tt.info, tt.ok); got != tt.want {
				t.Errorf("versionFromBuildInfo() = %q, want %q", got, tt.want)
			}
		})
	}
}
