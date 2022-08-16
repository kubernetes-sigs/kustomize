// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func TestCountCommand_files(t *testing.T) {
	d := t.TempDir()

	err := os.WriteFile(filepath.Join(d, "f1.yaml"), []byte(`
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`), 0600)
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetCountRunner("")
	r.Command.SetArgs([]string{d})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Contains(t, b.String(), "Deployment: 2\nService: 1\n") {
		return
	}
}

func TestCountSubPackages(t *testing.T) {
	var tests = []struct {
		name        string
		dataset     string
		packagePath string
		args        []string
		expected    string
	}{
		{
			name:    "count-recurse-subpackages",
			dataset: "dataset-without-setters",
			expected: `${baseDir}/

${baseDir}/mysql/
Deployment: 1

${baseDir}/mysql/storage/
Deployment: 1
`,
		},
		{
			name:        "count-top-level-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			args:        []string{"-R=false"},
			packagePath: "mysql",
			expected: `${baseDir}/mysql/
Deployment: 1
`,
		},
		{
			name:        "count-nested-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			packagePath: "mysql/storage",
			args:        []string{"-R=false"},
			expected: `${baseDir}/mysql/storage/
Deployment: 1
`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			openapi.ResetOpenAPI()
			defer openapi.ResetOpenAPI()
			sourceDir := filepath.Join("test", "testdata", test.dataset)
			baseDir := t.TempDir()
			copyutil.CopyDir(sourceDir, baseDir)
			runner := commands.GetCountRunner("")
			actual := &bytes.Buffer{}
			runner.Command.SetOut(actual)
			runner.Command.SetArgs(append([]string{filepath.Join(baseDir, test.packagePath)}, test.args...))
			err := runner.Command.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// normalize path format for windows
			actualNormalized := strings.ReplaceAll(
				strings.ReplaceAll(actual.String(), "\\", "/"),
				"//", "/")

			expected := strings.ReplaceAll(test.expected, "${baseDir}", baseDir)
			expectedNormalized := strings.ReplaceAll(expected, "\\", "/")
			if !assert.Equal(t, expectedNormalized, actualNormalized) {
				t.FailNow()
			}
		})
	}
}
