package main

import (
	"bytes"
	"log"
	"os/exec"
	"testing"
)

func TestLoadRepoManager(t *testing.T) {
	// Assuming gorepomod is installed
	path, err := exec.LookPath("gorepomod")
	if err != nil {
		log.Fatal(err)
	}
	var testCases = map[string]struct {
		cmd *exec.Cmd
	}{
		"withLocalFlag": {
			cmd: exec.Command(path, "list --local"),
		},
		"noLocalFlag": {
			cmd: exec.Command(path, "list"),
		},
	}

	for _, tc := range testCases {
		err := tc.cmd.Run()

		var out bytes.Buffer
		tc.cmd.Stdout = &out

		if err != nil {
			log.Fatal(err)
		}
	}
}
