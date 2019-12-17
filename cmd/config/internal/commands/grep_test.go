// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
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

	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
    config.kubernetes.io/index: '0'
    config.kubernetes.io/package: '.'
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
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
`, b.String()) {
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

	if !assert.Equal(t, `kind: Deployment
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
`, b.String()) {
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
	if !assert.Equal(t, `kind: Deployment
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
`, b.String()) {
		return
	}
}
