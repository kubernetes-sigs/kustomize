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

func TestCreateSubstitutionCommand(t *testing.T) {
	var tests = []struct {
		name              string
		inputOpenAPI      string
		input             string
		args              []string
		out               string
		expectedOpenAPI   string
		expectedResources string
	}{
		{
			name: "substitution replicas",
			args: []string{
				"image", "nginx:1.7.9", "--pattern", "IMAGE:TAG",
				"--value", "IMAGE=image", "--value", "TAG=tag"},
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
        image: nginx:1.7.9
      - name: sidecar
        image: sidecar:1.7.9
 `,
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
		},
		{
			name: "substitution and create setters 1",
			args: []string{
				"image", "something/nginx::1.7.9/nginxotherthing", "--pattern", "something/IMAGE::TAG/nginxotherthing",
				"--value", "IMAGE=image", "--value", "TAG=tag"},
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
        image: something/nginx::1.7.9/nginxotherthing
      - name: sidecar
        image: sidecar:1.7.9
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
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
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: 1.7.9
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: something/IMAGE::TAG/nginxotherthing
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
        image: something/nginx::1.7.9/nginxotherthing # {"$ref":"#/definitions/io.k8s.cli.substitutions.image"}
      - name: sidecar
        image: sidecar:1.7.9
 `,
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

			runner := commands.NewCreateSubstitutionRunner("")
			out := &bytes.Buffer{}
			runner.Command.SetOut(out)
			runner.Command.SetArgs(append([]string{r.Name()}, test.args...))
			err = runner.Command.Execute()
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
