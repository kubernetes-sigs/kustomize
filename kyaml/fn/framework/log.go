package framework

import (
	"fmt"
	"os"
)

// Log prints to stderr.
func Log(in ...interface{}) {
	fmt.Fprintln(os.Stderr, in...)
}

// Logf prints formatted messages to stderr.
func Logf(format string, in ...interface{}) {
	fmt.Fprintf(os.Stderr, format, in...)
}
