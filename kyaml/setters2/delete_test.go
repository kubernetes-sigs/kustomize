// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestDelete_Filter(t *testing.T) {
	var tests = []struct {
		name           string
		description    string
		setter         string
		input          string
		expectedOutput string
	}{
		{
			name:   "delete-replicas",
			setter: "replicas",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    replicas: 3 # {"$openapi":"replicas"}
spec:
  replicas: 3 # {"$openapi":"replicas"}
 `,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    replicas: 3
spec:
  replicas: 3
 `,
		},
		{
			name:   "delete-foo-annotation",
			setter: "foo",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: 3 # {"$openapi":"foo"}
 `,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: 3
 `,
		},
		{
			name:   "delete-replicas-enum",
			setter: "replicas",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 1 # {"$openapi":"replicas"}
 `,
			expectedOutput: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 1
 `,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// parse the input to be modified
			r, err := yaml.Parse(test.input)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// invoke the delete
			instance := &Delete{FieldName: test.setter}
			result, err := instance.Filter(r)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// compare the actual and expected output
			actual, err := result.String()
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			actual = strings.TrimSpace(actual)
			expected := strings.TrimSpace(test.expectedOutput)
			if !assert.Equal(t, expected, actual) {
				t.FailNow()
			}
		})
	}
}

var resourcefile2 = `apiVersion: resource.dev/v1alpha1
kind: resourcefile
metadata:
    name: hello-world-set
upstream:
    type: git
    git:
        commit: 5c1c019b59299a4f6c7edd1ff5ff54d720621bbe
        directory: /package-examples/helloworld-set
        ref: v0.1.0
packageMetadata:
    shortDescription: example package using setters
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: "2"
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "sometag"
`

func TestDelete_Filter2(t *testing.T) {
	path := filepath.Join(os.TempDir(), "resourcefile2")

	//write initial resourcefile to temp path
	err := ioutil.WriteFile(path, []byte(resourcefile2), 0666)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	//add a deleter definition
	dd := DeleterDefinition{
		Name: "image",
	}

	err = dd.DeleteFromFile(path)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.FailNow()
	}

	expected := `apiVersion: resource.dev/v1alpha1
kind: resourcefile
metadata:
  name: hello-world-set
upstream:
  type: git
  git:
    commit: 5c1c019b59299a4f6c7edd1ff5ff54d720621bbe
    directory: /package-examples/helloworld-set
    ref: v0.1.0
packageMetadata:
  shortDescription: example package using setters
openAPI:
  definitions:
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "sometag"
`
	assert.Equal(t, expected, string(b))
}
