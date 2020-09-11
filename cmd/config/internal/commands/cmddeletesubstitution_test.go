// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
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
			out: `deleted substitution "my.image"`,
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
			out: `deleted substitution "my-image-sub"`,
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
			err: `substitution "my-image-sub-not-present" does not exist`,
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
			err: `substitution "my-image-subst" is used in substitution "my-nested-subst", please delete the parent substitution first`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			openapi.ResetOpenAPI()
			defer openapi.ResetOpenAPI()

			baseDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.RemoveAll(baseDir)
			f := filepath.Join(baseDir, "Krmfile")
			err = ioutil.WriteFile(f, []byte(test.inputOpenAPI), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			r, err := ioutil.TempFile(baseDir, "k8s-cli-*.yaml")
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
			runner.Command.SetArgs(append([]string{baseDir}, test.args...))
			err = runner.Command.Execute()

			if test.err != "" {
				if !assert.NotNil(t, err) {
					t.FailNow()
				}
				assert.Equal(t, test.err, err.Error())
				return
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// normalize path format for windows
			actualNorm := strings.Replace(
				strings.Replace(out.String(), "\\", "/", -1),
				"//", "/", -1)
			expectedOut := strings.Replace(test.out, "${baseDir}", baseDir, -1)
			expectedNorm := strings.Replace(expectedOut, "\\", "/", -1)

			if !assert.Contains(t, strings.TrimSpace(actualNorm), expectedNorm) {
				t.FailNow()
			}

			actualOpenAPI, err := ioutil.ReadFile(f)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(test.expectedOpenAPI),
				strings.TrimSpace(string(actualOpenAPI))) {
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
		})
	}
}

func TestDeleteSubstitutionSubPackages(t *testing.T) {
	var tests = []struct {
		name        string
		dataset     string
		packagePath string
		args        []string
		expected    string
	}{
		{
			name:    "delete-substitution-recurse-subpackages",
			dataset: "dataset-with-setters",
			args:    []string{"image-tag", "-R"},
			expected: `${baseDir}/mysql/
deleted substitution "image-tag"

${baseDir}/mysql/nosetters/
substitution "image-tag" does not exist

${baseDir}/mysql/storage/
substitution "image-tag" does not exist
`,
		},
		{
			name:        "delete-setter-top-level-pkg-no-recurse-subpackages",
			dataset:     "dataset-with-setters",
			packagePath: "mysql",
			args:        []string{"image-tag"},
			expected: `${baseDir}/mysql/
deleted substitution "image-tag"
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
			baseDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			copyutil.CopyDir(sourceDir, baseDir)
			//defer os.RemoveAll(baseDir)
			runner := commands.NewDeleteSubstitutionRunner("")
			actual := &bytes.Buffer{}
			runner.Command.SetOut(actual)
			runner.Command.SetArgs(append([]string{filepath.Join(baseDir, test.packagePath)}, test.args...))
			err = runner.Command.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// normalize path format for windows
			actualNormalized := strings.Replace(
				strings.Replace(actual.String(), "\\", "/", -1),
				"//", "/", -1)

			expected := strings.Replace(test.expected, "${baseDir}", baseDir, -1)
			expectedNormalized := strings.Replace(expected, "\\", "/", -1)
			if !assert.Equal(t, expectedNormalized, actualNormalized) {
				t.FailNow()
			}
		})
	}
}
