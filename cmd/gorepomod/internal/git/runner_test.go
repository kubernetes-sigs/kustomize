// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestCheckDiffOnFile(t *testing.T) {
	defer t.Cleanup(cleanup)

	f, err := os.Create("./g.test")
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	defer f.Close()

	f, err = os.OpenFile("../../../../.gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	d := []byte("test\n")
	f.Write(d)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	defer f.Close()

	var testCases = map[string]struct {
		filePath string
		expected string
	}{
		"TestDiffNotFound": {
			filePath: "./g.test",
			expected: "",
		},
		"TestDiffFound": {
			filePath: "../../../../.gitignore",
			expected: `
			--- a/.gitignore
			+++ b/.gitignore			
			`,
		},
	}
	for n, tc := range testCases {
		gr := NewQuiet(".", true)
		actual, _ := gr.CheckDiffOnFile(tc.filePath)
		if n == "TestDiffNotFound" {
			if len(actual) > len(tc.expected) {
				t.Fatalf(
					"%s: for %s, expected %s, got %s",
					n, tc.filePath, tc.expected, actual)
			}
		} else {
			if strings.Contains(tc.expected, actual) {
				t.Fatalf(
					"%s: for %s, expected %s, got %s",
					n, tc.filePath, tc.expected, actual)
			}
		}
	}
}

func cleanup() {
	gr := NewQuiet(".", true)

	// Undo creation of g.test
	err := os.Remove("./g.test")
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	// Undo changes on .gitignore
	err = gr.runNoOut(noHarmDone, "restore", "../../../../.gitignore")
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
}
