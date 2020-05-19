// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package starlark

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFilter_Filter(t *testing.T) {
	var tests = []struct {
		name                   string
		input                  string
		functionConfig         string
		script                 string
		expected               string
		expectedFunctionConfig string
		env                    map[string]string
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

run(ctx.resource_list["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: bar
    config.kubernetes.io/path: 'deployment_nginx-deployment.yaml'
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
			name: "add_annotation_from_env",
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
def run(r):
  for resource in r:
    resource["metadata"]["annotations"]["foo"] = ctx.environment["ANNOTATION"]

run(ctx.resource_list["items"])
`,
			env: map[string]string{"ANNOTATION": "annotation-value"},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: annotation-value
    config.kubernetes.io/path: 'deployment_nginx-deployment.yaml'
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
			name: "add_annotation_from_open_api",
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
def run(r):
  for resource in r:
    resource["metadata"]["annotations"]["foo"] = ctx.open_api["definitions"]["io.k8s.api.apps.v1.Deployment"]["description"]

run(ctx.resource_list["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: Deployment enables declarative updates for Pods and ReplicaSets.
    config.kubernetes.io/path: 'deployment_nginx-deployment.yaml'
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

run(ctx.resource_list["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: bar
    config.kubernetes.io/path: 'deployment_nginx-deployment.yaml'
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
			name: "delete_annotation",
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
    resource["metadata"]["annotations"].pop("foo")

run(ctx.resource_list["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/path: 'deployment_nginx-deployment.yaml'
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

run(ctx.resource_list["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
  annotations:
    foo: bar
    config.kubernetes.io/path: 'deployment_nginx-deployment-1.yaml'
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
    config.kubernetes.io/path: 'deployment_nginx-deployment-2.yaml'
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
run(ctx.resource_list["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
  annotations:
    config.kubernetes.io/path: 'deployment_nginx-deployment-1.yaml'
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
    config.kubernetes.io/path: 'deployment_nginx-deployment-2.yaml'
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
run(ctx.resource_list["items"])
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-1
  annotations:
    config.kubernetes.io/path: 'deployment_nginx-deployment-1.yaml'
`,
		},
		{
			name: "functionConfig",
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
			functionConfig: `
kind: Script
spec:
  value: "hello world"
`,
			script: `
# set the foo annotation on each resource
def run(r, an):
  for resource in r:
    resource["metadata"]["annotations"]["foo"] = an

an = ctx.resource_list["functionConfig"]["spec"]["value"]
run(ctx.resource_list["items"], an)
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: hello world
    config.kubernetes.io/path: 'deployment_nginx-deployment.yaml'
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
			expectedFunctionConfig: `
kind: Script
spec:
  value: "hello world"
`,
		},

		{
			name: "functionConfig_update_functionConfig",
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
			functionConfig: `
kind: Script
spec:
  value: "hello world"
`,
			script: `
# set the foo annotation on each resource
def run(r, an):
  for resource in r:
    resource["metadata"]["annotations"]["foo"] = an

an = ctx.resource_list["functionConfig"]["spec"]["value"]
run(ctx.resource_list["items"], an)
ctx.resource_list["functionConfig"]["spec"]["value"] = "updated"
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: hello world
    config.kubernetes.io/path: 'deployment_nginx-deployment.yaml'
spec:
  template:
    spec:
      containers:
      - name: nginx
        # head comment
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`,
			expectedFunctionConfig: `
kind: Script
spec:
  value: "hello world"
`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.env {
				os.Setenv(k, v)
			}
			f := &Filter{Name: test.name, Program: test.script}

			if test.functionConfig != "" {
				fc, err := yaml.Parse(test.functionConfig)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				f.FunctionConfig = fc
			}

			r := &kio.ByteReader{Reader: bytes.NewBufferString(test.input)}
			o := &bytes.Buffer{}
			w := &kio.ByteWriter{Writer: o}
			p := kio.Pipeline{
				Inputs:  []kio.Reader{r},
				Filters: []kio.Filter{f},
				Outputs: []kio.Writer{w},
			}
			err := p.Execute()
			if !assert.NoError(t, err) {
				if e, ok := err.(*errors.Error); ok {
					fmt.Fprintf(os.Stderr, "%s\n", e.Stack())
				}
				t.FailNow()
			}
			if !assert.Equal(t, strings.TrimSpace(test.expected), strings.TrimSpace(o.String())) {
				t.FailNow()
			}

			if test.expectedFunctionConfig != "" {
				if !assert.Equal(t,
					strings.TrimSpace(test.expectedFunctionConfig),
					strings.TrimSpace(f.FunctionConfig.MustString())) {
					t.FailNow()
				}
			}
		})
	}
}
