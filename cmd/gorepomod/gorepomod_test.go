// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommand(t *testing.T) {
	// Assuming gorepomod is installed
	var testCases = map[string]struct {
		isFork   bool
		cmd      string
		expected string
	}{
		"upstreamWithLocalFlag": {
			isFork: false,
			cmd:    "cd ../.. && gorepomod list --local",
		},
		"upstreamWithNoLocalFlag": {
			isFork: false,
			cmd:    "cd ../.. && gorepomod list",
		},
		"forkWithLocalFlag": {
			isFork: true,
			cmd:    "cd ../.. && gorepomod list --local",
		},
		"forkWithNoLocalFlag": {
			isFork: true,
			cmd:    "cd ../.. && gorepomod list",
		},
	}

	for _, tc := range testCases {
		bash, err := exec.LookPath("bash")
		if err != nil {
			t.Error("bash not found")
		}
		out, err := exec.Command(bash, "-c", tc.cmd).Output()
		if err != nil {
			require.Error(t, err, "exit status 1")
		}
		assert.Greater(t, len(string(out)), 1)
	}
}

func TestPinCommand(t *testing.T) {
	// Assuming gorepomod is installed
	var testCases = map[string]struct {
		isFork bool
		cmd    string
	}{
		"upstreamWithLocalFlag": {
			isFork: false,
			cmd:    "cd ../.. && gorepomod pin kyaml --local",
		},
		"upstreamWithNoLocalFlag": {
			isFork: false,
			cmd:    "cd ../.. && gorepomod pin kyaml",
		},
		"forkWithLocalFlag": {
			isFork: true,
			cmd:    "cd ../.. && gorepomod pin kyaml --local",
		},
		"forkWithNoLocalFlag": {
			isFork: true,
			cmd:    "cd ../.. && gorepomod pin kyaml",
		},
	}

	for _, tc := range testCases {
		bash, err := exec.LookPath("bash")
		if err != nil {
			t.Error("bash not found")
		}
		out, err := exec.Command(bash, "-c", tc.cmd).Output()
		if err != nil {
			require.Error(t, err, "exit status 1")
		}
		assert.Greater(t, len(string(out)), 1)
	}
}
