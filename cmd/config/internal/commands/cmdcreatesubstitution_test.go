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

func TestCreateSubstitutionCommand(t *testing.T) {
	var tests = []struct {
		name              string
		inputOpenAPI      string
		input             string
		args              []string
		out               string
		expectedOpenAPI   string
		expectedResources string
		err               string
	}{
		{
			name: "substitution replicas",
			args: []string{
				"my-image-subst", "--field-value", "nginx:1.7.9", "--pattern", "${my-image-setter}:${my-tag-setter}"},
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
    io.k8s.cli.setters.my-image-setter:
      x-k8s-cli:
        setter:
          name: my-image-setter
          value: "nginx"
    io.k8s.cli.setters.my-tag-setter:
      x-k8s-cli:
        setter:
          name: my-tag-setter
          value: "1.7.9"
 `,
			out: `created substitution "my-image-subst"`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.my-image-setter:
      x-k8s-cli:
        setter:
          name: my-image-setter
          value: "nginx"
    io.k8s.cli.setters.my-tag-setter:
      x-k8s-cli:
        setter:
          name: my-tag-setter
          value: "1.7.9"
    io.k8s.cli.substitutions.my-image-subst:
      x-k8s-cli:
        substitution:
          name: my-image-subst
          pattern: ${my-image-setter}:${my-tag-setter}
          values:
          - marker: ${my-image-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-image-setter'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
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
        image: nginx:1.7.9 # {"$openapi":"my-image-subst"}
      - name: sidecar
        image: sidecar:1.7.9
 `,
		},
		{
			name: "error if substitution with same name exists",
			args: []string{"my-image", "--field-value", "some:image", "--pattern", "some:${image}"},
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.substitutions.my-image:
      x-k8s-cli:
        substitution:
          name: my-image
          pattern: something/${my-image-setter}::${my-tag-setter}/nginxotherthing
          values:
          - marker: ${my-image-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-image-setter'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
 `,
			err: `substitution with name "my-image" already exists`,
		},
		{
			name: "error if setter with same name exists",
			args: []string{
				"my-image", "--field-value", "nginx:1.7.9", "--pattern", "${my-image-setter}:${my-tag-setter}"},
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
 `,
			err: `setter with name "my-image" already exists, substitution and setter can't have same name`,
		},
		{
			name: "substitution and create setters 1",
			args: []string{
				"my-image-subst", "--field-value", "something/nginx::1.7.9/nginxotherthing", "--pattern", "something/${my-image-setter}::${my-tag-setter}/nginxotherthing"},
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
			out: `created substitution "my-image-subst"`,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.substitutions.my-image-subst:
      x-k8s-cli:
        substitution:
          name: my-image-subst
          pattern: something/${my-image-setter}::${my-tag-setter}/nginxotherthing
          values:
          - marker: ${my-image-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-image-setter'
          - marker: ${my-tag-setter}
            ref: '#/definitions/io.k8s.cli.setters.my-tag-setter'
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
        image: something/nginx::1.7.9/nginxotherthing # {"$openapi":"my-image-subst"}
      - name: sidecar
        image: sidecar:1.7.9
 `,
		},
		{
			name: "nested substitution",
			args: []string{
				"my-nested-subst", "--field-value", "something/nginx::1.7.9/nginxotherthing",
				"--pattern", "something/${my-image-subst}/${my-other-setter}"},
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
        image: nginx::1.7.9 # {"$openapi":"my-image-subst"}
 `,
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
 `,
			out: `created substitution "my-nested-subst"`,
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
		},
		{
			name: "substitution with non-existing setter with same name",
			args: []string{
				"foo", "--field-value", "prefix-1234", "--pattern", "prefix-${foo}"},
			input: `
apiVersion: test/v1
kind: Foo
metadata:
  name: foo
spec:
  setterVal: 1234
  substVal: prefix-1234
 `,
			inputOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedOpenAPI: `
apiVersion: v1alpha1
kind: Example
 `,
			expectedResources: `
apiVersion: test/v1
kind: Foo
metadata:
  name: foo
spec:
  setterVal: 1234
  substVal: prefix-1234
        	            	
 `,
			err: `setters must have different name than the substitution: foo`,
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

			runner := commands.NewCreateSubstitutionRunner("")
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

			expectedOut := strings.Replace(test.out, "${baseDir}", baseDir, -1)
			expectedNormalized := strings.Replace(expectedOut, "\\", "/", -1)
			// normalize path format for windows
			actualNormalized := strings.Replace(
				strings.Replace(out.String(), "\\", "/", -1),
				"//", "/", -1)

			if !assert.Contains(t, actualNormalized, expectedNormalized) {
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

			actualOpenAPI, err := ioutil.ReadFile(f)
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

func TestCreateSubstSubPackages(t *testing.T) {
	var tests = []struct {
		name        string
		dataset     string
		packagePath string
		args        []string
		expected    string
	}{
		{
			name:    "create-subst-recurse-subpackages",
			dataset: "dataset-without-setters",
			args:    []string{"image-tag", "--field-value", "mysql:1.7.9", "--pattern", "${image}:${tag}", "-R"},
			expected: `${baseDir}/mysql/
created substitution "image-tag"

${baseDir}/mysql/storage/
created substitution "image-tag"
`,
		},
		{
			name:        "create-subst-top-level-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			packagePath: "mysql",
			args:        []string{"image-tag", "--field-value", "mysql:1.7.9", "--pattern", "${image}:${tag}"},
			expected: `${baseDir}/mysql/
created substitution "image-tag"`,
		},
		{
			name:        "create-subst-nested-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			packagePath: "mysql/storage",
			args:        []string{"image-tag", "--field-value", "storage:1.7.9", "--pattern", "${image}:${tag}"},
			expected: `${baseDir}/mysql/storage/
created substitution "image-tag"`,
		},
		{
			name:        "create-subst-already-exists",
			dataset:     "dataset-with-setters",
			packagePath: "mysql",
			args:        []string{"image-tag", "--field-value", "mysql:1.7.9", "--pattern", "${image}:${tag}", "-R"},
			expected: `${baseDir}/mysql/
substitution with name "image-tag" already exists

${baseDir}/mysql/nosetters/
created substitution "image-tag"

${baseDir}/mysql/storage/
created substitution "image-tag"`,
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
			defer os.RemoveAll(baseDir)
			runner := commands.NewCreateSubstitutionRunner("")
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
			if !assert.Equal(t, strings.TrimSpace(expectedNormalized), strings.TrimSpace(actualNormalized)) {
				t.FailNow()
			}
		})
	}
}
