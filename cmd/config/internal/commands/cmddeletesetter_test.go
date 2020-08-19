// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func TestDeleteSetterCommand(t *testing.T) {
	var tests = []struct {
		name              string
		input             string
		args              []string
		schema            string
		out               string
		inputOpenAPI      string
		expectedOpenAPI   string
		expectedResources string
		err               string
	}{
		{
			name: "delete replicas",
			args: []string{"replicas-setter"},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi" : "replicas-setter"}
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas-setter:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas-setter
          value: "3"
          setBy: me
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		},
		{
			name: "delete only one setter",
			args: []string{"replicas-setter"},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi" : "replicas-setter"}
  foo: nginx # {"$openapi" : "image"}
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas-setter:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas-setter
          value: "3"
          setBy: me
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: nginx
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: nginx
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo: nginx # {"$openapi" : "image"}
 `,
		},

		{
			name: "delete array setter",
			args: []string{"list"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      items:
        type: string
      maxItems: 3
      type: array
      description: hello world
      x-k8s-cli:
        setter:
          name: list
          value: ""
          listValues:
          - a
          - b
          - c
          setBy: me
 `,
			input: `
apiVersion: example.com/v1beta1
kind: Example1
spec:
  list: # {"$openapi":"list"}
  - "a"
  - "b"
  - "c"
---
apiVersion: example.com/v1beta1
kind: Example2
spec:
  list: # {"$openapi":"list2"}
  - "a"
  - "b"
  - "c"
 `,
			expectedResources: `
apiVersion: example.com/v1beta1
kind: Example1
spec:
  list:
  - "a"
  - "b"
  - "c"
---
apiVersion: example.com/v1beta1
kind: Example2
spec:
  list: # {"$openapi":"list2"}
  - "a"
  - "b"
  - "c"
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
`,
		},

		{
			name: "delete non exist setter error",
			args: []string{"image"},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi" : "replicas-setter"}
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas-setter:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas-setter
          value: "3"
          setBy: me
`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas-setter:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas-setter
          value: "3"
          setBy: me
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi" : "replicas-setter"}
 `,
			err: `setter with name image does not exist`,
		},
		{
			name: "delete setter used in substitution error",
			args: []string{"image-name"},
			input: `
apiVersion: apps/v1
kind: Deployment
 `,
			inputOpenAPI: `
openAPI:
  definitions:
    io.k8s.cli.setters.image-name:
      x-k8s-cli:
        setter:
          name: image-name
          value: "nginx"
    io.k8s.cli.setters.image-tag:
      x-k8s-cli:
        setter:
          name: image-tag
          value: "1.8.1"
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE_NAME:IMAGE_TAG
          values:
          - marker: "IMAGE_NAME"
            ref: "#/definitions/io.k8s.cli.setters.image-name"
          - marker: "IMAGE_TAG"
            ref: "#/definitions/io.k8s.cli.setters.image-tag"
`,
			expectedOpenAPI: `
openAPI:
  definitions:
    io.k8s.cli.setters.image-name:
      x-k8s-cli:
        setter:
          name: image-name
          value: "nginx"
    io.k8s.cli.setters.image-tag:
      x-k8s-cli:
        setter:
          name: image-tag
          value: "1.8.1"
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE_NAME:IMAGE_TAG
          values:
          - marker: "IMAGE_NAME"
            ref: "#/definitions/io.k8s.cli.setters.image-name"
          - marker: "IMAGE_TAG"
            ref: "#/definitions/io.k8s.cli.setters.image-tag"
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
 `,
			err: `setter is used in substitution image, please delete the parent substitution first`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			openapi.ResetOpenAPI()
			defer openapi.ResetOpenAPI()

			f, err := ioutil.TempFile("", "k8s-cli-")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.Remove(f.Name())

			err = ioutil.WriteFile(f.Name(), []byte(test.inputOpenAPI), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			old := ext.GetOpenAPIFile
			defer func() { ext.GetOpenAPIFile = old }()
			ext.GetOpenAPIFile = func(args []string) (s string, err error) {
				return f.Name(), nil
			}

			r, err := ioutil.TempFile("", "k8s-cli-*.yaml")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.Remove(r.Name())
			err = ioutil.WriteFile(r.Name(), []byte(test.input), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			runner := commands.NewDeleteSetterRunner("")
			out := &bytes.Buffer{}
			runner.Command.SetOut(out)
			runner.Command.SetArgs(append([]string{r.Name()}, test.args...))
			err = runner.Command.Execute()
			if test.err != "" {
				if !assert.NotNil(t, err) {
					t.FailNow()
				} else {
					assert.Equal(t, err.Error(), test.err)
					return
				}
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t, test.out, out.String()) {
				t.FailNow()
			}

			actualResources, err := ioutil.ReadFile(r.Name())
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(test.expectedResources),
				strings.TrimSpace(string(actualResources))) {
				t.FailNow()
			}

			actualOpenAPI, err := ioutil.ReadFile(f.Name())
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(test.expectedOpenAPI),
				strings.TrimSpace(string(actualOpenAPI))) {
				t.FailNow()
			}
		})
	}
}
