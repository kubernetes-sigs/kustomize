// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestAnnotateCommand(t *testing.T) {
	var tests = []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "single value",
			args:     []string{"--kv", "a=b"},
			expected: expectedSingleValue,
		},
		{
			name:     "multi value",
			args:     []string{"--kv", "a=b", "--kv", "c=d"},
			expected: expectedMultiValue,
		},
		{
			name:     "filter kind",
			args:     []string{"--kv", "a=b", "--kind", "Service"},
			expected: expectedFilterKindService,
		},
		{
			name:     "filter apiVersion",
			args:     []string{"--kv", "a=b", "--apiVersion", "v1"},
			expected: expectedFilterApiVersionV1,
		},
		{
			name:     "filter name",
			args:     []string{"--kv", "a=b", "--name", "bar"},
			expected: expectedFilterNameBar,
		},
		{
			name:     "filter namespace",
			args:     []string{"--kv", "a=b", "--namespace", "bar"},
			expected: expectedFilterNamespaceBar,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			d := initTestDir(t)
			defer os.RemoveAll(d)

			a := NewAnnotateRunner("")
			a.Command.SetArgs(append([]string{d}, tt.args...))
			a.Command.SilenceUsage = true
			a.Command.SilenceErrors = true

			err := a.Command.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actual := &bytes.Buffer{}
			err = kio.Pipeline{
				Inputs:  []kio.Reader{kio.LocalPackageReader{PackagePath: d}},
				Outputs: []kio.Writer{kio.ByteWriter{Writer: actual, KeepReaderAnnotations: true}},
			}.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(tt.expected),
				strings.TrimSpace(actual.String())) {
				t.FailNow()
			}
		})
	}
}

func initTestDir(t *testing.T) string {
	d, err := ioutil.TempDir("", "kustomize-annotate-test")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = ioutil.WriteFile(filepath.Join(d, "f1.yaml"), []byte(f1Input), 0600)
	if !assert.NoError(t, err) {
		defer os.RemoveAll(d)
		t.FailNow()
	}
	err = ioutil.WriteFile(filepath.Join(d, "f2.yaml"), []byte(f2Input), 0600)
	if !assert.NoError(t, err) {
		defer os.RemoveAll(d)
		t.FailNow()
	}
	return d
}

var (
	f1Input = `
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
`

	f2Input = `
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
  namespace: bar
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
  namespace: foo
spec:
  replicas: 3
`

	expectedSingleValue = `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
    a: 'b'
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
    a: 'b'
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
---
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
    a: 'b'
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: bar
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
    a: 'b'
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: foo
spec:
  replicas: 3
`

	expectedMultiValue = `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
    a: 'b'
    c: 'd'
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
    a: 'b'
    c: 'd'
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
---
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
    a: 'b'
    c: 'd'
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: bar
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
    a: 'b'
    c: 'd'
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: foo
spec:
  replicas: 3
`

	expectedFilterKindService = `kind: Deployment
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
    a: 'b'
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: nginx
---
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: bar
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
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: foo
spec:
  replicas: 3
`

	expectedFilterApiVersionV1 = `kind: Deployment
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
---
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
    a: 'b'
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: bar
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
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: foo
spec:
  replicas: 3
`

	expectedFilterNameBar = `kind: Deployment
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
---
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: bar
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
    a: 'b'
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: foo
spec:
  replicas: 3
`

	expectedFilterNamespaceBar = `kind: Deployment
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
---
apiVersion: v1
kind: Abstraction
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/example/reconciler:v1
  annotations:
    config.kubernetes.io/local-config: "true"
    a: 'b'
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: bar
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
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'f2.yaml'
  namespace: foo
spec:
  replicas: 3
`
)
