// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
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
	t.Helper()
	d := t.TempDir()
	err := ioutil.WriteFile(filepath.Join(d, "f1.yaml"), []byte(f1Input), 0600)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = ioutil.WriteFile(filepath.Join(d, "f2.yaml"), []byte(f2Input), 0600)
	if !assert.NoError(t, err) {
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f1.yaml'
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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'f2.yaml'
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
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'f2.yaml'
  namespace: foo
spec:
  replicas: 3
`
)

func TestAnnotateSubPackages(t *testing.T) {
	var tests = []struct {
		name        string
		dataset     string
		packagePath string
		args        []string
		expected    string
	}{
		{
			name:    "annotate-recurse-subpackages",
			dataset: "dataset-without-setters",
			args:    []string{"--kv", "foo=bar", "-R"},
			expected: `${baseDir}/
added annotations in the package

${baseDir}/mysql/
added annotations in the package

${baseDir}/mysql/storage/
added annotations in the package
`,
		},
		{
			name:        "annotate-top-level-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			packagePath: "mysql",
			args:        []string{"--kv", "foo=bar"},
			expected: `${baseDir}/mysql/
added annotations in the package
`,
		},
		{
			name:        "annotate-nested-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			packagePath: "mysql/storage",
			args:        []string{"--kv", "foo=bar"},
			expected: `${baseDir}/mysql/storage/
added annotations in the package
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
			runner := NewAnnotateRunner("")
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
			if !assert.Contains(t, actualNormalized, expectedNormalized) {
				t.FailNow()
			}
		})
	}
}
