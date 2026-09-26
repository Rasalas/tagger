package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/rasalas/tagger/internal/term"
	"github.com/spf13/cobra"
)

var version = "0.0.1-dev"

var rootCmd = &cobra.Command{
	Use:           "tagger",
	Short:         "Semantic Git tagging based on Conventional Commits",
	Long:          "Tagger reads commits since the last tag, determines the next semver bump, and creates the tag.",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       buildVersion(),
}

func init() {
	rootCmd.SetOut(termWriter{})
	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return configErr(err)
	})
}

// termWriter resolves term.W on each write so callers can replace it after init.
type termWriter struct{}

func (termWriter) Write(p []byte) (int, error) {
	return term.W.Write(p)
}

func noArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.NoArgs(cmd, args); err != nil {
		return configErr(err)
	}
	return nil
}

// Execute runs the root command.
// Exit codes: 0 = success, 1 = runtime errors, 2 = configuration errors.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		var cfgErr *ConfigError
		code := 1
		if errors.As(err, &cfgErr) {
			code = 2
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(code)
	}
}

// ConfigError marks a configuration problem (invalid flags or tags) so it
// exits with code 2 instead of 1.
type ConfigError struct {
	Err error
}

func (e *ConfigError) Error() string { return e.Err.Error() }

func (e *ConfigError) Unwrap() error { return e.Err }

func configErr(err error) error { return &ConfigError{Err: err} }

func buildVersion() string {
	if version != "" && version != "0.0.1-dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	return versionFromBuildInfo(info, ok)
}

func versionFromBuildInfo(info *debug.BuildInfo, ok bool) string {
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}
