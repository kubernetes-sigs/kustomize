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

func TestDeleteSubstitutionCommand(t *testing.T) {
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
			name: "delete only subst if setter has same name - long ref",
			args: []string{"my.image"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my.image:
      x-k8s-cli:
        setter:
          name: my.image
          value: "nginx"
    io.k8s.cli.setters.my-tag:
      x-k8s-cli:
        setter:
          name: my-tag
          value: "1.7.9"
    io.k8s.cli.substitutions.my.image:
      x-k8s-cli:
        substitution:
          name: my.image
          pattern: ${my.image}:${my-tag}
          values:
          - marker: ${my.image}
            ref: '#/definitions/io.k8s.cli.setters.my.image'
          - marker: ${my-tag}
            ref: '#/definitions/io.k8s.cli.setters.my-tag'
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
  replicas: 3 # {"$openapi":"replicas"}
  template:
    spec:
      containers:
      - name: nginx # {"$ref":"#/definitions/io.k8s.cli.setters.my.image"}
        image: nginx:1.7.9 # {"$ref":"#/definitions/io.k8s.cli.substitutions.my.image"}
      - name: sidecar
        image: nginx:1.7.9 # {"$ref":"#/definitions/io.k8s.cli.substitutions.my.image"}
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
  template:
    spec:
      containers:
      - name: nginx # {"$ref":"#/definitions/io.k8s.cli.setters.my.image"}
        image: nginx:1.7.9
      - name: sidecar
        image: nginx:1.7.9
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my.image:
      x-k8s-cli:
        setter:
          name: my.image
          value: "nginx"
    io.k8s.cli.setters.my-tag:
      x-k8s-cli:
        setter:
          name: my-tag
          value: "1.7.9"
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
		},
		{
			name: "delete subst - short ref",
			args: []string{"my-image-sub"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image:
      x-k8s-cli:
        setter:
          name: my-image
          value: "nginx"
    io.k8s.cli.setters.my-tag:
      x-k8s-cli:
        setter:
          name: my-tag
          value: "1.7.9"
    io.k8s.cli.substitutions.my-image-sub:
      x-k8s-cli:
        substitution:
          name: my-image-sub
          pattern: ${my-image}:${my-tag}
          values:
          - marker: ${my-image}
            ref: '#/definitions/io.k8s.cli.setters.my-image'
          - marker: ${my-tag}
            ref: '#/definitions/io.k8s.cli.setters.my-tag'
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
  replicas: 3 # {"$openapi":"replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$openapi":"my-image-sub"}
      - name: sidecar
        image: nginx:1.7.9 # {"$openapi":"my-image-sub"}
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
      - name: sidecar
        image: nginx:1.7.9
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image:
      x-k8s-cli:
        setter:
          name: my-image
          value: "nginx"
    io.k8s.cli.setters.my-tag:
      x-k8s-cli:
        setter:
          name: my-tag
          value: "1.7.9"
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
		},
		{
			name: "error if subst doesn't exist",
			args: []string{"my-image-sub-not-present"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image:
      x-k8s-cli:
        setter:
          name: my-image
          value: "nginx"
    io.k8s.cli.setters.my-tag:
      x-k8s-cli:
        setter:
          name: my-tag
          value: "1.7.9"
    io.k8s.cli.substitutions.my-image-sub:
      x-k8s-cli:
        substitution:
          name: my-image-sub
          pattern: ${my-image}:${my-tag}
          values:
          - marker: ${my-image}
            ref: '#/definitions/io.k8s.cli.setters.my-image'
          - marker: ${my-tag}
            ref: '#/definitions/io.k8s.cli.setters.my-tag'
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
  replicas: 3 # {"$openapi":"replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$openapi":"my-image-sub"}
      - name: sidecar
        image: nginx:1.7.9 # {"$openapi":"my-image-sub"}
 `,
			expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$openapi":"replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
      - name: sidecar
        image: nginx:1.7.9
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image:
      x-k8s-cli:
        setter:
          name: my-image
          value: "nginx"
    io.k8s.cli.setters.my-tag:
      x-k8s-cli:
        setter:
          name: my-tag
          value: "1.7.9"
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			err: "substitution with name my-image-sub-not-present does not exist",
		},

		{
			name: "substitution referenced by other substitution",
			args: []string{"my-image-subst"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image-setter:
      x-k8s-cli:
        setter:
          name: my-image-setter
          value: nginx
    io.k8s.cli.setters.my-tag-setter:
      x-k8s-cli:
        setter:
          name: my-tag-setter
          value: 1.7.9
    io.k8s.cli.substitutions.my-image-subst:
      x-k8s-cli:
        substitution:
          name: my-image-subst
          pattern: ${my-image-setter}::${my-tag-setter}
          values:
          - marker: ${my-image-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-image-setter'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
    io.k8s.cli.substitutions.my-nested-subst:
      x-k8s-cli:
        substitution:
          name: my-nested-subst
          pattern: something/${my-image-subst}/${my-other-setter}
          values:
          - marker: ${my-image-subst}
            ref: '#/definitions/io.k8s.cli.substitutions.my-image-subst'
          - marker: ${my-other-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-other-setter'
    io.k8s.cli.setters.my-other-setter:
      x-k8s-cli:
        setter:
          name: my-other-setter
          value: nginxotherthing
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
        image: something/nginx::1.7.9/nginxotherthing # {"$openapi":"my-nested-subst"}
      - name: sidecar
        image: nginx::1.7.9 # {"$openapi":"my-image-subst"}
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image-setter:
      x-k8s-cli:
        setter:
          name: my-image-setter
          value: nginx
    io.k8s.cli.setters.my-tag-setter:
      x-k8s-cli:
        setter:
          name: my-tag-setter
          value: 1.7.9
    io.k8s.cli.substitutions.my-image-subst:
      x-k8s-cli:
        substitution:
          name: my-image-subst
          pattern: ${my-image-setter}::${my-tag-setter}
          values:
          - marker: ${my-image-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-image-setter'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
    io.k8s.cli.substitutions.my-nested-subst:
      x-k8s-cli:
        substitution:
          name: my-nested-subst
          pattern: something/${my-image-subst}/${my-other-setter}
          values:
          - marker: ${my-image-subst}
            ref: '#/definitions/io.k8s.cli.substitutions.my-image-subst'
          - marker: ${my-other-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-other-setter'
    io.k8s.cli.setters.my-other-setter:
      x-k8s-cli:
        setter:
          name: my-other-setter
          value: nginxotherthing
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
        image: something/nginx::1.7.9/nginxotherthing # {"$openapi":"my-nested-subst"}
      - name: sidecar
        image: nginx::1.7.9 # {"$openapi":"my-image-subst"}
 `,
			err: "substitution is used in substitution my-nested-subst, please delete the parent substitution first",
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

			runner := commands.NewDeleteSubstitutionRunner("")
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
				return
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
