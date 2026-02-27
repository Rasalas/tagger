package cmd

import (
	"fmt"
	"os"

	"github.com/rasalas/tagger/internal/term"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "cli-template",
	Short:         "A brief description of your tool",
	Long:          "A longer description of what your tool does.",
	RunE:          run,
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

func run(cmd *cobra.Command, args []string) error {
	term.Header("cli-template")
	term.Info("Hello! Replace this with your tool's logic.")
	term.Pass("Setup complete")
	fmt.Fprintln(term.W)
	return nil
}
