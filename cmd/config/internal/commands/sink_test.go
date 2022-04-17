// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
)

func TestSinkCommand(t *testing.T) {
	d := t.TempDir()

	r := commands.GetSinkRunner("")
	r.Command.SetIn(bytes.NewBufferString(`apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- kind: Deployment
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
- kind: Service
  metadata:
    name: foo
    annotations:
      app: nginx
      config.kubernetes.io/index: '1'
      config.kubernetes.io/path: 'f1.yaml'
  spec:
    selector:
      app: nginx
- apiVersion: v1
  kind: Abstraction
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/function: |
        container:
          image: gcr.io/example/reconciler:v1
      config.kubernetes.io/local-config: "true"
      config.kubernetes.io/index: '0'
      config.kubernetes.io/path: 'f2.yaml'
  spec:
    replicas: 3
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: nginx
    name: bar
    annotations:
      app: nginx
      config.kubernetes.io/index: '1'
      config.kubernetes.io/path: 'f2.yaml'
  spec:
    replicas: 3
`))
	r.Command.SetArgs([]string{d})
	if !assert.NoError(t, r.Command.Execute()) {
		t.FailNow()
	}

	actual, err := ioutil.ReadFile(filepath.Join(d, "f1.yaml"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	expected := `kind: Deployment
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
`
	if !assert.Equal(t, expected, string(actual)) {
		t.FailNow()
	}

	actual, err = ioutil.ReadFile(filepath.Join(d, "f2.yaml"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	expected = `apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/example/reconciler:v1
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 3
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`
	if !assert.Equal(t, expected, string(actual)) {
		t.FailNow()
	}
}

func TestSinkCommandJSON(t *testing.T) {
	d := t.TempDir()

	r := commands.GetSinkRunner("")
	r.Command.SetIn(bytes.NewBufferString(`apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- {"kind": "Deployment", "metadata": {"labels": {"app": "nginx2"}, "name": "foo",
    "annotations": {"app": "nginx2", config.kubernetes.io/index: '0',
      config.kubernetes.io/path: 'f1.json'}}, "spec": {"replicas": 1}}
`))
	r.Command.SetArgs([]string{d})
	if !assert.NoError(t, r.Command.Execute()) {
		t.FailNow()
	}

	actual, err := ioutil.ReadFile(filepath.Join(d, "f1.json"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	expected := `{
  "kind": "Deployment",
  "metadata": {
    "annotations": {
      "app": "nginx2"
    },
    "labels": {
      "app": "nginx2"
    },
    "name": "foo"
  },
  "spec": {
    "replicas": 1
  }
}
`
	if !assert.Equal(t, expected, string(actual)) {
		t.FailNow()
	}
}

func TestSinkCommand_Stdout(t *testing.T) {
	// fmt the files
	out := &bytes.Buffer{}
	r := commands.GetSinkRunner("")
	r.Command.SetIn(bytes.NewBufferString(`apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- kind: Deployment
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
- kind: Service
  metadata:
    name: foo
    annotations:
      app: nginx
      config.kubernetes.io/index: '1'
      config.kubernetes.io/path: 'f1.yaml'
  spec:
    selector:
      app: nginx
- apiVersion: v1
  kind: Abstraction
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/function: |
        container:
          image: gcr.io/example/reconciler:v1
      config.kubernetes.io/local-config: "true"
      config.kubernetes.io/index: '0'
      config.kubernetes.io/path: 'f2.yaml'
  spec:
    replicas: 3
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: nginx
    name: bar
    annotations:
      app: nginx
      config.kubernetes.io/index: '1'
      config.kubernetes.io/path: 'f2.yaml'
  spec:
    replicas: 3
`))

	r.Command.SetOut(out)
	r.Command.SetArgs([]string{})
	if !assert.NoError(t, r.Command.Execute()) {
		t.FailNow()
	}

	expected := `kind: Deployment
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
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/example/reconciler:v1
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 3
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
spec:
  replicas: 3
`
	if !assert.Equal(t, expected, out.String()) {
		t.FailNow()
	}
}

func TestSinkCommandJSON_Stdout(t *testing.T) {
	// fmt the files
	out := &bytes.Buffer{}
	r := commands.GetSinkRunner("")
	r.Command.SetIn(bytes.NewBufferString(`apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- {"kind": "Deployment", "metadata": {"labels": {"app": "nginx2"}, "name": "foo",
    "annotations": {"app": "nginx2", config.kubernetes.io/index: '0',
      config.kubernetes.io/path: 'f1.json'}}, "spec": {"replicas": 1}}
`))

	r.Command.SetOut(out)
	r.Command.SetArgs([]string{})
	if !assert.NoError(t, r.Command.Execute()) {
		t.FailNow()
	}

	expected := `{
  "kind": "Deployment",
  "metadata": {
    "annotations": {
      "app": "nginx2"
    },
    "labels": {
      "app": "nginx2"
    },
    "name": "foo"
  },
  "spec": {
    "replicas": 1
  }
}
`
	if !assert.Equal(t, expected, out.String()) {
		println(out.String())
		t.FailNow()
	}
}
