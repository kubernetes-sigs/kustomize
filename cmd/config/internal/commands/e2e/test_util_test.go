// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/testutil"
)

func TestMain(m *testing.M) {
	d := build()
	defer os.RemoveAll(d)
	os.Exit(m.Run())
}

type test struct {
	name           string
	args           []string
	files          map[string]string
	expectedFiles  map[string]string
	expectedErr    string
	expectedStdOut string
}

func runTests(t *testing.T, tests []test) {
	dir := build()
	bin := filepath.Join(dir, kyamlBin)

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			dataDir, err := ioutil.TempDir("", "kustomize-test-data-")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.RemoveAll(dataDir)
			os.Chdir(dataDir)

			// write the input
			for path, data := range tt.files {
				err := ioutil.WriteFile(path, []byte(data), 0600)
				testutil.AssertNoError(t, err)
			}

			cmd := exec.Command(bin, tt.args...)
			cmd.Dir = dataDir
			var stdErr, stdOut bytes.Buffer
			cmd.Stdout = &stdOut
			cmd.Stderr = &stdErr
			cmd.Env = os.Environ()

			err = cmd.Run()
			if tt.expectedErr != "" {
				if !assert.Contains(t, stdErr.String(), tt.expectedErr, stdErr.String()) {
					t.FailNow()
				}
				return
			}
			testutil.AssertNoError(t, err, stdErr.String(), stdOut.String())

			if tt.expectedStdOut != "" {
				if !assert.Equal(t, strings.TrimSpace(stdOut.String()), strings.TrimSpace(tt.expectedStdOut)) {
					t.FailNow()
				}
			}

			for path, data := range tt.expectedFiles {
				b, err := ioutil.ReadFile(path)
				testutil.AssertNoError(t, err, stdErr.String())

				if !assert.Equal(t, strings.TrimSpace(data), strings.TrimSpace(string(b)), stdErr.String()) {
					t.FailNow()
				}
			}
		})
	}
}
