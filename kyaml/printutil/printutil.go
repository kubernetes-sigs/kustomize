// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package printutil

import (
	"fmt"
	"io"
)

// GreenPrintf formats according to a format specifier and prints the resulting string
// in green color to the writer
func GreenPrintf(w io.Writer, format string, args ...string) {
	msg := fmt.Sprintf(format, args)
	fmt.Fprintf(w, "\033[1;32m%s\033[0m", msg)
}

// WarnPrintf formats according to a format specifier and prints the resulting string
// in yellow color to the writer
func WarnPrintf(w io.Writer, format string, args ...string) {
	msg := fmt.Sprintf(format, args)
	fmt.Fprintf(w, "\033[1;33m%s\033[0m", msg)
}
