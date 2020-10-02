// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
)

// TestGrepCommand_files verifies grep reads the files and filters them
func TestGrepCommand_files(t *testing.T) {
	d, err := ioutil.TempDir("", "kustomize-kyaml-test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(d)

	err = ioutil.WriteFile(filepath.Join(d, "f1.yaml"), []byte(`
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
	err = ioutil.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`kind: Deployment
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
	r := commands.GetGrepRunner("")
	r.Command.SetArgs([]string{"metadata.name=foo", d})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Contains(t, b.String(), `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
`) {
		return
	}
}

// TestCmd_stdin verifies the grep command reads stdin if no files are provided
func TestGrepCmd_stdin(t *testing.T) {
	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetGrepRunner("")
	r.Command.SetArgs([]string{"metadata.name=foo"})
	r.Command.SetOut(b)
	r.Command.SetIn(bytes.NewBufferString(`
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
---
kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`))
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Contains(t, b.String(), `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
    config.kubernetes.io/index: '0'
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
    config.kubernetes.io/index: '1'
spec:
  selector:
    app: nginx
`) {
		return
	}
}

// TestGrepCmd_errInputs verifies the grep command errors on invalid matches
func TestGrepCmd_errInputs(t *testing.T) {
	b := &bytes.Buffer{}
	r := commands.GetGrepRunner("")
	r.Command.SetArgs([]string{"metadata.name=foo=bar"})
	r.Command.SetOut(b)
	r.Command.SetIn(bytes.NewBufferString(`
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
`))
	err := r.Command.Execute()
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "ambiguous match")

	// fmt the files
	b = &bytes.Buffer{}
	r = commands.GetGrepRunner("")
	r.Command.SetArgs([]string{"spec.template.spec.containers[a[b=c].image=foo"})
	r.Command.SetOut(b)
	r.Command.SetIn(bytes.NewBufferString(`
kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
`))
	err = r.Command.Execute()
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "unrecognized path element:")
}

// TestGrepCommand_escapeDots verifies the grep command correctly escapes '\.' in inputs
func TestGrepCommand_escapeDots(t *testing.T) {
	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetGrepRunner("")
	r.Command.SetArgs([]string{"spec.template.spec.containers[name=nginx].image=nginx:1\\.7\\.9",
		"--annotate=false"})
	r.Command.SetOut(b)
	r.Command.SetIn(bytes.NewBufferString(`
kind: Deployment
metadata:
  labels:
    app: nginx1.8
  name: nginx1.8
  annotations:
    app: nginx1.8
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.8.1
---
kind: Deployment
metadata:
  labels:
    app: nginx1.7
  name: nginx1.7
  annotations:
    app: nginx1.7
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
`))
	err := r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Contains(t, b.String(), `kind: Deployment
metadata:
  labels:
    app: nginx1.7
  name: nginx1.7
  annotations:
    app: nginx1.7
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
`) {
		return
	}
}

func TestGrepSubPackages(t *testing.T) {
	var tests = []struct {
		name        string
		dataset     string
		packagePath string
		args        []string
		expected    string
	}{
		{
			name:    "grep-recurse-subpackages",
			dataset: "dataset-without-setters",
			args:    []string{"kind=Deployment"},
			expected: `
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: myspace
  name: mysql-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'deployment.yaml'
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: mysql
        image: mysql:1.7.9
---
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: myspace
  name: storage-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'deployment.yaml'
spec:
  replicas: 4
  template:
    spec:
      containers:
      - name: storage
        image: storage:1.7.7
`,
		},
		{
			name:        "grep-top-level-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			args:        []string{"kind=Deployment", "-R=false"},
			packagePath: "mysql",
			expected: `# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: myspace
  name: mysql-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'deployment.yaml'
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: mysql
        image: mysql:1.7.9
`,
		},
		{
			name:        "grep-nested-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			packagePath: "mysql/storage",
			args:        []string{"kind=Deployment", "-R=false"},
			expected: `# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: myspace
  name: storage-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'deployment.yaml'
spec:
  replicas: 4
  template:
    spec:
      containers:
      - name: storage
        image: storage:1.7.7
`,
		},
		{
			name:    "grep-recurse-subpackages-no-result",
			dataset: "dataset-without-setters",
			args:    []string{"kind=ConfigMap"},
			expected: `

`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			sourceDir := filepath.Join("test", "testdata", test.dataset)
			baseDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			copyutil.CopyDir(sourceDir, baseDir)
			defer os.RemoveAll(baseDir)
			runner := commands.GetGrepRunner("")
			actual := &bytes.Buffer{}
			runner.Command.SetOut(actual)
			runner.Command.SetArgs(append(test.args, filepath.Join(baseDir, test.packagePath)))
			err = runner.Command.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// normalize path format for windows
			actualNormalized := strings.Replace(
				strings.Replace(actual.String(), "\\", "/", -1),
				"//", "/", -1)

			expected := strings.Replace(test.expected, "${baseDir}", baseDir, -1)
			expectedNormalized := strings.Replace(expected, "\\", "/", -1)
			if !assert.Equal(t, expectedNormalized, actualNormalized) {
				t.FailNow()
			}
		})
	}
}
