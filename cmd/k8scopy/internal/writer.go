// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type writer struct {
	root string
	f    *os.File
}

func newWriter(toDir, name string) (*writer, error) {
	if err := os.MkdirAll(toDir, 0755); err != nil {
		log.Printf("unable to create directory: %s", toDir)
		return nil, err
	}
	n := filepath.Join(toDir, name)
	f, err := os.Create(n)
	if err != nil {
		return nil, fmt.Errorf("unable to create `%s`; %v", n, err)
	}
	return &writer{root: toDir, f: f}, nil
}

func (w *writer) close() {
	w.f.Close()
}

func (w *writer) write(line string) {
	if _, err := w.f.WriteString(line + "\n"); err != nil {
		log.Printf("Trouble writing: %s", line)
		log.Fatalf("Error: %s", err)
	}
}
