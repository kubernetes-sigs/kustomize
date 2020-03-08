// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package starlark

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestStarlarkFilter_Filter(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		script   string
		expected string
	}{
		{
			name: "add_annotation",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
			script: `
# set the foo annotation on each resource
def run(r):
  for resource in r:
    resource["metadata"]["annotations"]["foo"] = "bar"

run(resourceList["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: bar
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
		},
		{
			name: "update_annotation",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: baz
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
			script: `
# set the foo annotation on each resource
def run(r):
  for resource in r:
    resource["metadata"]["annotations"]["foo"] = "bar"

run(resourceList["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: bar
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
		},
		{
			name: "update_annotation_multiple",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
  annotations:
    foo: baz
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-2
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
			script: `
# set the foo annotation on each resource
def run(r):
  for resource in r:
    resource["metadata"]["annotations"]["foo"] = "bar"

run(resourceList["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
  annotations:
    foo: bar
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-2
  annotations:
    foo: bar
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}`,
		},
		{
			name: "add_resource",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
			script: `
def run(r):
  d = {
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "nginx-deployment-2",
  },
}
  r.append(d)
run(resourceList["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
		},
		{
			name: "remove_resource",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-2
`,
			script: `
def run(r):
  r.pop()
run(resourceList["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			r := &kio.ByteReader{Reader: bytes.NewBufferString(test.input)}
			o := &bytes.Buffer{}
			w := &kio.ByteWriter{Writer: o}
			f := &Filter{Name: test.name, Program: test.script}
			p := kio.Pipeline{
				Inputs:  []kio.Reader{r},
				Filters: []kio.Filter{f},
				Outputs: []kio.Writer{w},
			}
			err := p.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t, strings.TrimSpace(test.expected), strings.TrimSpace(o.String())) {
				t.FailNow()
			}
		})
	}
}
