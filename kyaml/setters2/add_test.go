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

func TestAdd_Filter(t *testing.T) {
	var tests = []struct {
		name     string
		add      Add
		input    string
		expected string
		err      string
	}{
		{
			name: "add-replicas",
			add: Add{
				FieldValue: "3",
				Ref:        "#/definitions/io.k8s.cli.setters.replicas",
			},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name: "add-replicas-annotations",
			add: Add{
				FieldValue: "3",
				Ref:        "#/definitions/io.k8s.cli.setters.replicas",
			},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    something: 3
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    something: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name: "add-replicas-name",
			add: Add{
				FieldValue: "3",
				FieldName:  "replicas",
				Ref:        "#/definitions/io.k8s.cli.setters.replicas",
			},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    something: 3
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    something: 3
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name: "add-replicas-2x",
			add: Add{
				FieldValue: "3",
				FieldName:  "replicas",
				Ref:        "#/definitions/io.k8s.cli.setters.replicas",
			},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    replicas: 3
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name: "add-replicas-1x",
			add: Add{
				FieldValue: "3",
				FieldName:  "spec.replicas",
				Ref:        "#/definitions/io.k8s.cli.setters.replicas",
			},
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    replicas: 3
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    replicas: 3
spec:
  replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name: "add-replicas-error",
			add: Add{
				Ref: "#/definitions/io.k8s.cli.setters.replicas",
			},
			err: "must specify either fieldName or fieldValue",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
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

			// invoke add
			result, err := test.add.Filter(r)
			if test.err != "" {
				if !assert.Equal(t, test.err, err.Error()) {
					t.FailNow()
				}
				return
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// compare the actual and expected output
			actual, err := result.String()
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			actual = strings.TrimSpace(actual)
			expected := strings.TrimSpace(test.expected)
			if !assert.Equal(t, expected, actual) {
				t.FailNow()
			}
		})
	}
}

var resourcefile = `apiVersion: resource.dev/v1alpha1
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
    shortDescription: example package using setters`

func TestAdd_Filter2(t *testing.T) {
	path := filepath.Join(os.TempDir(), "resourcefile")

	//write initial resourcefile to temp path
	err := ioutil.WriteFile(path, []byte(resourcefile), 0666)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	//add a setter definition
	sd := SetterDefinition{
		Name:  "image",
		Value: "1",
	}

	err = sd.AddToFile(path)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// update setter definition
	sd2 := SetterDefinition{
		Name:  "image",
		Value: "2",
	}

	err = sd2.AddToFile(path)

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
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: "2"
`
	assert.Equal(t, expected, string(b))
}

func TestAddUpdateSubstitution(t *testing.T) {
	path := filepath.Join(os.TempDir(), "resourcefile")

	//write initial resourcefile to temp path
	err := ioutil.WriteFile(path, []byte(resourcefile), 0666)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	value1 := Value{
		Marker: "IMAGE_NAME",
		Ref:    "#/definitions/io.k8s.cli.setters.image-name",
	}

	value2 := Value{
		Marker: "IMAGE_TAG",
		Ref:    "#/definitions/io.k8s.cli.setters.image-tag",
	}

	values := []Value{value1, value2}

	//add a setter definition
	subd := SubstitutionDefinition{
		Name:    "image",
		Pattern: "IMAGE_NAME:IMAGE_TAG",
		Values:  values,
	}

	err = subd.AddToFile(path)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// update setter definition
	subd2 := SubstitutionDefinition{
		Name:    "image",
		Pattern: "IMAGE_NAME:IMAGE_TAG2",
	}

	err = subd2.AddToFile(path)

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
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE_NAME:IMAGE_TAG2
          values: []
`
	assert.Equal(t, expected, string(b))
}
