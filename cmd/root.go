package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "tagger",
	Short:         "Semantic Git tagging based on Conventional Commits",
	Long:          "Tagger reads commits since the last tag, determines the next semver bump, and creates the tag.",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       "0.0.1-dev",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if exitErr, ok := err.(*exitError); ok {
			os.Exit(exitErr.code)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}

type exitError struct {
	code int
}

func (e *exitError) Error() string {
	return fmt.Sprintf("exit %d", e.code)
}
