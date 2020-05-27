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

func TestSetCommand(t *testing.T) {
	var tests = []struct {
		name              string
		inputOpenAPI      string
		input             string
		args              []string
		out               string
		expectedOpenAPI   string
		expectedResources string
		errMsg            string
	}{
		{
			name: "set replicas",
			args: []string{"--name", "replicas", "--value", "4", "--description", "hi there", "--set-by", "pw"},
			out:  "set 1 fields\n",
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hi there
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
          setBy: pw
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name: "set replicas no description",
			args: []string{"--name", "replicas", "--value", "4"},
			out:  "set 1 fields\n",
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name: "set replicas no description",
			args: []string{"--name", "name", "--value", "-test"},
			out:  "set 1 fields\n",
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.name:
      description: hello world
      x-k8s-cli:
        setter:
          name: name
          value: "t"
          setBy: me
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  name: t # {"$ref":"#/definitions/io.k8s.cli.setters.name"}
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.name:
      description: hello world
      x-k8s-cli:
        setter:
          name: name
          value: "-test"
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  name: -test # {"$ref":"#/definitions/io.k8s.cli.setters.name"}
 `,
		},
		{
			name: "set image",
			args: []string{"--name", "tag", "--value", "1.8.1"},
			out:  "set 1 fields\n",
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref":"#/definitions/io.k8s.cli.substitutions.image"}
      - name: sidecar
        image: sidecar:1.7.9
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
          value: "nginx"
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "1.8.1"
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'

 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.8.1 # {"$ref":"#/definitions/io.k8s.cli.substitutions.image"}
      - name: sidecar
        image: sidecar:1.7.9
`,
		},

		{
			name: "validate openAPI number",
			args: []string{"--name", "replicas", "--value", "four"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      type: number
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      type: number
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
			errMsg: "replicas in body must be of type number",
		},

		{
			name: "validate openAPI string maxLength",
			args: []string{"--name", "name", "--value", "wordpress"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.name:
      type: string
      maxLength: 5
      description: hello world
      x-k8s-cli:
        setter:
          name: name
          value: nginx
          setBy: me
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx # {"$ref":"#/definitions/io.k8s.cli.setters.name"}
spec:
  replicas: 3
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.name:
      type: string
      maxLength: 5
      description: hello world
      x-k8s-cli:
        setter:
          name: name
          value: nginx
          setBy: me
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx # {"$ref":"#/definitions/io.k8s.cli.setters.name"}
spec:
  replicas: 3
 `,
			errMsg: "name in body should be at most 5 chars long",
		},

		{
			name: "validate substitution",
			args: []string{"--name", "tag", "--value", "1.8.1"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
    io.k8s.cli.setters.tag:
      type: string
      minLength: 6
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref":"#/definitions/io.k8s.cli.substitutions.image"}
      - name: sidecar
        image: sidecar:1.7.9
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
          value: "nginx"
    io.k8s.cli.setters.tag:
      type: string
      minLength: 6
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE:TAG
          values:
          - marker: IMAGE
            ref: '#/definitions/io.k8s.cli.setters.image'
          - marker: TAG
            ref: '#/definitions/io.k8s.cli.setters.tag'

 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref":"#/definitions/io.k8s.cli.substitutions.image"}
      - name: sidecar
        image: sidecar:1.7.9
`,
			errMsg: "tag in body should be at least 6 chars long",
		},

		{
			name: "validate openAPI list values",
			args: []string{"--name", "list", "--list-value", "10, hi, true"},
			inputOpenAPI: `
kind: Kptfile
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      type: array
      maxItems: 2
      items:
        type: integer
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - 0
 `,
			input: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$ref":"#/definitions/io.k8s.cli.setters.list"}
  - 0
 `,
			expectedOpenAPI: `
kind: Kptfile
openAPI:
  definitions:
    io.k8s.cli.setters.list:
      type: array
      maxItems: 2
      items:
        type: integer
      x-k8s-cli:
        setter:
          name: list
          listValues:
          - 0
 `,
			expectedResources: `
apiVersion: example.com/v1beta1
kind: Example
spec:
  list: # {"$ref":"#/definitions/io.k8s.cli.setters.list"}
  - 0
 `,
			errMsg: `list in body must be of type integer: "string"
list in body must be of type integer: "boolean"
list in body should have at most 2 items`,
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

			runner := commands.NewSetRunner("")
			out := &bytes.Buffer{}
			runner.Command.SetOut(out)
			runner.Command.SetArgs(append([]string{r.Name()}, test.args...))
			err = runner.Command.Execute()
			if test.errMsg != "" {
				if !assert.NotNil(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), test.errMsg) {
					t.FailNow()
				}
			}

			if test.errMsg == "" && !assert.NoError(t, err) {
				t.FailNow()
			}

			if test.errMsg == "" && !assert.Equal(t, test.out, out.String()) {
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
