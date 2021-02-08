// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func RunNoOutputCommand(n string, args ...string) {
	o := RunGetOutputCommand(n, args...)
	if len(o) > 0 {
		log.Fatalf("unexpected output: %q", o)
	}
}

func RunGetOutputCommand(n string, args ...string) string {
	cmd := exec.Command(n, args...)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		fmt.Printf("err: %q\n", errBuf.String())
		log.Fatal(err)
	}
	return strings.TrimSpace(outBuf.String())
}
