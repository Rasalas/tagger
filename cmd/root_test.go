package cmd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/rasalas/tagger/internal/git"
	"github.com/rasalas/tagger/internal/term"
	"github.com/spf13/cobra"
)

func TestExecuteProcess(t *testing.T) {
	if os.Getenv("TAGGER_TEST_EXECUTE_PROCESS") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			rootCmd.SetArgs(os.Args[i+1:])
			break
		}
	}
	git.Default = &mockGit{latestErr: errors.New("repository unavailable")}
	Execute()
}

func TestExecuteExitCodes(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		code    int
		message string
	}{
		{"unknown flag", []string{"bump", "--unknown"}, 2, "unknown flag"},
		{"missing flag value", []string{"bump", "--prefix"}, 2, "needs an argument"},
		{"invalid boolean", []string{"bump", "--yes=nope"}, 2, "invalid argument"},
		{"conflicting flags", []string{"status", "--major", "--minor"}, 2, "only one"},
		{"positional argument", []string{"bump", "dry-run", "--yes"}, 2, "unknown command"},
		{"runtime error", []string{"status"}, 1, "repository unavailable"},
		{"successful help", []string{"--help"}, 0, "Usage:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"-test.run=^TestExecuteProcess$", "--"}, tt.args...)
			child := exec.Command(os.Args[0], args...)
			child.Env = append(os.Environ(), "TAGGER_TEST_EXECUTE_PROCESS=1")
			out, err := child.CombinedOutput()
			code := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("start child: %v", err)
				}
				code = exitErr.ExitCode()
			}
			if code != tt.code || !strings.Contains(string(out), tt.message) {
				t.Fatalf("exit=%d output=%q, want exit=%d containing %q", code, out, tt.code, tt.message)
			}
		})
	}
}

func TestCobraOutputUsesCurrentTermWriter(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{"root help", []string{"--help"}, "Usage:"},
		{"bump help", []string{"bump", "--help"}, "tagger bump [flags]"},
		{"version", []string{"--version"}, "tagger version"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			original := term.W
			t.Cleanup(func() {
				term.W = original
				rootCmd.SetArgs(nil)
				for _, command := range []*cobra.Command{rootCmd, bumpCmd, statusCmd} {
					for _, flag := range []string{"help", "version"} {
						if command.Flags().Lookup(flag) != nil {
							if err := command.Flags().Set(flag, "false"); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
			})
			for range 2 {
				var output bytes.Buffer
				term.W = &output
				rootCmd.SetArgs(tt.args)
				if err := rootCmd.Execute(); err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(output.String(), tt.want) {
					t.Fatalf("term.W captured %q, want %q", output.String(), tt.want)
				}
			}
		})
	}
}

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
