package term

import (
	"fmt"
	"io"
	"os"
)

// W is the output writer, defaults to stdout.
var W io.Writer = os.Stdout

// Header prints a section header.
func Header(title string) {
	fmt.Fprintf(W, "\n  %s%s%s\n\n", Bold, title, Reset)
}

// Pass prints a passing check line.
func Pass(msg string) {
	fmt.Fprintf(W, "  %s✓%s %s\n", Green, Reset, msg)
}

// Fail prints a failing check line.
func Fail(msg string) {
	fmt.Fprintf(W, "  %s✗%s %s\n", Red, Reset, msg)
}

// FailDetail prints a failing check with extra detail.
func FailDetail(key, detail string) {
	fmt.Fprintf(W, "  %s✗%s %s %s— %s%s\n", Red, Reset, key, Dim, detail, Reset)
}

// Warn prints a warning line.
func Warn(msg string) {
	fmt.Fprintf(W, "  %s!%s %s\n", Yellow, Reset, msg)
}

// WarnDetail prints a warning with extra detail.
func WarnDetail(key, detail string) {
	fmt.Fprintf(W, "  %s!%s %s %s— %s%s\n", Yellow, Reset, key, Dim, detail, Reset)
}

// Info prints an informational line.
func Info(msg string) {
	fmt.Fprintf(W, "  %s%s%s\n", Dim, msg, Reset)
}
