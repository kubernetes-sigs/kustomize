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

// TODO(pwittrock): write tests for reading / writing ResourceLists

func TestCmd_files(t *testing.T) {
	d, err := ioutil.TempDir("", "kustomize-cat-test")
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
	err = ioutil.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
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
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetCatRunner("")
	r.Command.SetArgs([]string{d})
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
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  labels:
    app: nginx
  annotations:
    app: nginx
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f2.yaml'
spec:
  replicas: 3
`, b.String()) {
		return
	}
}

func TestCmd_filesWithReconcilers(t *testing.T) {
	d, err := ioutil.TempDir("", "kustomize-cat-test")
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
	err = ioutil.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/image:version
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 3
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
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetCatRunner("")
	r.Command.SetArgs([]string{d, "--include-local"})
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
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
---
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f2.yaml'
  configFn:
    container:
      image: gcr.io/example/image:version
spec:
  replicas: 3
---
kind: Deployment
metadata:
  labels:
    app: nginx
  name: bar
  annotations:
    app: nginx
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f2.yaml'
spec:
  replicas: 3
`, b.String()) {
		return
	}
}

func TestCmd_filesWithoutNonReconcilers(t *testing.T) {
	d, err := ioutil.TempDir("", "kustomize-cat-test")
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
	err = ioutil.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  annotations:
    config.kubernetes.io/local-config: "true"
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
spec:
  replicas: 3
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
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetCatRunner("")
	r.Command.SetArgs([]string{d, "--include-local", "--exclude-non-local"})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, `apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f2.yaml'
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
spec:
  replicas: 3
`, b.String()) {
		return
	}
}

func TestCmd_outputTruncateFile(t *testing.T) {
	d, err := ioutil.TempDir("", "kustomize-cat-test")
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
	err = ioutil.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
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
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	f, err := ioutil.TempFile("", "kustomize-cat-test-dest")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(d)

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetCatRunner("")
	r.Command.SetArgs([]string{d, "--dest", f.Name()})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, ``, b.String()) {
		return
	}

	actual, err := ioutil.ReadFile(f.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
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
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  labels:
    app: nginx
  annotations:
    app: nginx
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f2.yaml'
spec:
  replicas: 3
`, string(actual)) {
		return
	}
}

func TestCmd_outputCreateFile(t *testing.T) {
	d, err := ioutil.TempDir("", "kustomize-cat-test")
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
	err = ioutil.WriteFile(filepath.Join(d, "f2.yaml"), []byte(`
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
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
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	d2, err := ioutil.TempDir("", "kustomize-cat-test-dest")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(d2)
	f := filepath.Join(d2, "output.yaml")

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetCatRunner("")
	r.Command.SetArgs([]string{d, "--dest", f})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	if !assert.Equal(t, ``, b.String()) {
		return
	}

	actual, err := ioutil.ReadFile(f)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
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
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  labels:
    app: nginx
  annotations:
    app: nginx
    config.kubernetes.io/package: '.'
    config.kubernetes.io/path: 'f2.yaml'
spec:
  replicas: 3
`, string(actual)) {
		return
	}
}
