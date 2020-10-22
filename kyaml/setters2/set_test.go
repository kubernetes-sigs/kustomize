// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSet_Filter(t *testing.T) {
	var tests = []struct {
		name        string
		description string
		setter      string
		openapi     string
		input       string
		expected    string
	}{
		{
			name:   "set-replicas",
			setter: "replicas",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name:        "set-foo-type",
			description: "if a type is specified for a setter, ensure the field is of provided type",
			setter:      "foo",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.foo:
      x-k8s-cli:
        setter:
          name: foo
          value: "4"
      type: integer
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
		},
		{
			name:        "set-foo-type-float",
			description: "if a type is specified for a setter, ensure the field is of provided type",
			setter:      "foo",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.foo:
      x-k8s-cli:
        setter:
          name: foo
          value: "4.0"
      type: number
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: 4.0 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
		},
		{
			name:        "set-foo-no-type",
			description: "if a type is not specified for a setter or k8s schema, keep existing quoting",
			setter:      "foo",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.foo:
      x-k8s-cli:
        setter:
          name: foo
          value: "4"
 `,
			input: `
apiVersion: custom/v1
kind: Example
metadata:
  name: nginx-deployment
  annotations:
    foo: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
			expected: `
apiVersion: custom/v1
kind: Example
metadata:
  name: nginx-deployment
  annotations:
    foo: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
		},
		{
			name:   "set-replicas-enum",
			setter: "replicas",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "medium"
          enumValues:
            small: "1"
            medium: "5"
            large: "50"
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 1 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 5 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name:   "set-replicas-enum-large",
			setter: "replicas",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "large"
          enumValues:
            small: "1"
            medium: "5"
            large: "50"
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 1 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 50 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `,
		},
		{
			name:   "set-arg",
			setter: "arg1",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
    io.k8s.cli.setters.arg1:
      x-k8s-cli:
        setter:
          name: arg1
          value: "some value"
 `,
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
        args:
        - a
        - b # {"$ref": "#/definitions/io.k8s.cli.setters.arg1"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        args:
        - a
        - some value # {"$ref": "#/definitions/io.k8s.cli.setters.arg1"}`,
		},
		{
			name:   "substitute-image-tag",
			setter: "image-tag",
			openapi: `
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
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
 `,
		},
		{
			name:   "substitute-image-name-enum",
			setter: "image-tag",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.image-name:
      x-k8s-cli:
        setter:
          name: image-name
          value: "helloworld"
          enumValues:
            nginx: gcr.io/nginx
            helloworld: us.gcr.io/helloworld
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
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: us.gcr.io/helloworld:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
 `,
		},
		{
			name:   "substitute-annotation",
			setter: "project",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.project:
      x-k8s-cli:
        setter:
          name: project
          value: "a"
    io.k8s.cli.setters.location:
      x-k8s-cli:
        setter:
          name: location
          value: "b"
    io.k8s.cli.setters.cluster:
      x-k8s-cli:
        setter:
          name: cluster
          value: "c"
    io.k8s.cli.substitutions.key:
      x-k8s-cli:
        substitution:
          name: key
          pattern: https://container.googleapis.com/v1/projects/PROJECT/locations/LOCATION/clusters/CLUSTER
          values:
          - marker: "PROJECT"
            ref: "#/definitions/io.k8s.cli.setters.project"
          - marker: "LOCATION"
            ref: "#/definitions/io.k8s.cli.setters.location"
          - marker: "CLUSTER"
            ref: "#/definitions/io.k8s.cli.setters.cluster"
`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    key: 'https://container.googleapis.com/v1/projects/a/locations/a/clusters/a' # {"$ref": "#/definitions/io.k8s.cli.substitutions.key"}
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    key: 'https://container.googleapis.com/v1/projects/a/locations/b/clusters/c' # {"$ref": "#/definitions/io.k8s.cli.substitutions.key"}
`,
		},
		{
			name:   "substitute-not-match-setter",
			setter: "not-real",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.project:
      x-k8s-cli:
        setter:
          name: project
          value: "a"
    io.k8s.cli.setters.location:
      x-k8s-cli:
        setter:
          name: location
          value: "b"
    io.k8s.cli.setters.cluster:
      x-k8s-cli:
        setter:
          name: cluster
          value: "c"
    io.k8s.cli.substitutions.key:
      x-k8s-cli:
        substitution:
          name: key
          pattern: https://container.googleapis.com/v1/projects/PROJECT/locations/LOCATION/clusters/CLUSTER
          values:
          - marker: "PROJECT"
            ref: "#/definitions/io.k8s.cli.setters.project"
          - marker: "LOCATION"
            ref: "#/definitions/io.k8s.cli.setters.location"
          - marker: "CLUSTER"
            ref: "#/definitions/io.k8s.cli.setters.cluster"
`,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    key: 'https://container.googleapis.com/v1/projects/a/locations/a/clusters/a' # {"$ref": "#/definitions/io.k8s.cli.substitutions.key"}
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    key: 'https://container.googleapis.com/v1/projects/a/locations/a/clusters/a' # {"$ref": "#/definitions/io.k8s.cli.substitutions.key"}
`,
		},
		{
			name:   "substitute-image-name",
			setter: "image-name",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.image-name:
      x-k8s-cli:
        setter:
          name: image-name
          value: "foo"
    io.k8s.cli.setters.image-tag:
      x-k8s-cli:
        setter:
          name: image-tag
          value: "1.7.9"
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
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: foo:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
 `,
		},
		{
			name:   "substitute-substring",
			setter: "image-tag",
			openapi: `
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
        image: a:a # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
 `,
		},
		{
			name:   "set-args-list",
			setter: "args",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.args:
      x-k8s-cli:
        type: array
        setter:
          name: args
          listValues: ["1", "2", "3"]
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  # {"$ref": "#/definitions/io.k8s.cli.setters.args"}
  replicas: []
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  # {"$ref": "#/definitions/io.k8s.cli.setters.args"}
  replicas:
  - "1"
  - "2"
  - "3"
 `,
		},
		{
			name:   "set-args-list-replace",
			setter: "args",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.args:
      x-k8s-cli:
        type: array
        setter:
          name: args
          listValues: ["1", "2", "3"]
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  # {"$ref": "#/definitions/io.k8s.cli.setters.args"}
  replicas: ["4", "5"]
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  # {"$ref": "#/definitions/io.k8s.cli.setters.args"}
  replicas:
  - "1"
  - "2"
  - "3"
 `,
		},
		{
			name:        "set-with-invalid-type-int",
			description: "if a type is set to int instead of integer, we accept it",
			setter:      "foo",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.foo:
      x-k8s-cli:
        setter:
          name: foo
          value: "4"
      type: int
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
		},
		{
			name:        "set-with-invalid-type-bool",
			description: "if a type is set to bool instead of boolean, we accept it",
			setter:      "foo",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.foo:
      x-k8s-cli:
        setter:
          name: foo
          value: "true"
      type: bool
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: false # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    foo: true # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			defer openapi.ResetOpenAPI()
			initSchema(t, test.openapi)

			// parse the input to be modified
			r, err := yaml.Parse(test.input)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// invoke the setter
			instance := &Set{Name: test.setter}
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
			expected := strings.TrimSpace(test.expected)
			if !assert.Equal(t, expected, actual) {
				t.FailNow()
			}
		})
	}
}

func TestSet_SetAll(t *testing.T) {
	var tests = []struct {
		name        string
		description string
		setter      string
		openapi     string
		input       []string
		expected    []string
	}{
		{
			name:   "set-replicas-same-file",
			setter: "replicas",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
 `,
			input: []string{`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'cluster.yaml'
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment2
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'cluster.yaml'
spec:
  replicas: 10
 `},
			expected: []string{`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'cluster.yaml'
spec:
  replicas: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment2
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'cluster.yaml'
spec:
  replicas: 10
 `},
		},
		{
			name:   "set-replicas-different-file",
			setter: "replicas",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
 `,
			input: []string{`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'cluster.yaml'
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment2
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'another_cluster.yaml'
spec:
  replicas: 10
 `},
			expected: []string{`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'cluster.yaml'
spec:
  replicas: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `},
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			defer openapi.ResetOpenAPI()
			initSchema(t, test.openapi)

			// parse the input to be modified
			var inputNodes []*yaml.RNode
			for _, s := range test.input {
				r, err := yaml.Parse(s)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				inputNodes = append(inputNodes, r)
			}

			// invoke the setter
			instance := &Set{Name: test.setter}
			result, err := SetAll(instance).Filter(inputNodes)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// compare the actual and expected output
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t, len(result), len(test.expected)) {
				t.FailNow()
			}

			for i := range result {
				actual, _ := result[i].String()
				actual = strings.TrimSpace(actual)
				expected := strings.TrimSpace(test.expected[i])
				if !assert.Equal(t, expected, actual) {
					t.FailNow()
				}
			}
		})
	}
}

// initSchema initializes the openAPI with the definitions from s
func initSchema(t *testing.T, s string) {
	// parse out the schema from the input openAPI
	y, err := yaml.Parse(s)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	// get the field containing the openAPI
	f := y.Field("openAPI")
	if !assert.NotNil(t, f) {
		t.FailNow()
	}
	defs, err := f.Value.String()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// convert the yaml openAPI to an interface{}
	// which can be marshalled into json
	var o interface{}
	err = yaml.Unmarshal([]byte(defs), &o)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// convert the interface{} into a json string
	j, err := json.Marshal(o)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// reset the openAPI to clear existing definitions
	openapi.ResetOpenAPI()

	// add the json schema to the global schema
	err = openapi.AddSchema(j)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func TestSetOpenAPI_Filter(t *testing.T) {
	var tests = []struct {
		name        string
		setter      string
		value       string
		values      []string
		input       string
		expected    string
		description string
		setBy       string
		err         string
		isSet       bool
	}{
		{
			name:   "set-replicas",
			setter: "replicas",
			value:  "3",
			isSet:  true,
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
          required: true
          isSet: true
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
 `,
			expected: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          required: true
          isSet: true
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
`,
		},
		{
			name:   "set-annotation-quoted",
			setter: "replicas",
			value:  "3",
			isSet:  true,
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: 4
 `,
			expected: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          isSet: true
`,
		},
		{
			name:        "set-replicas-description",
			setter:      "replicas",
			value:       "3",
			description: "hello world",
			isSet:       true,
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
 `,
			expected: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          isSet: true
      description: hello world
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
`,
		},
		{
			name:   "set-replicas-set-by",
			setter: "replicas",
			value:  "3",
			setBy:  "carl",
			isSet:  true,
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
 `,
			expected: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: carl
          isSet: true
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
`,
		},
		{
			name:   "set-replicas-set-by-empty",
			setter: "replicas",
			value:  "3",
			isSet:  true,
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
          setBy: "package-default"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"
          setBy: "package-default"
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
          setBy: "package-default"
 `,
			expected: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
          setBy: "package-default"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          isSet: true
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
          setBy: "package-default"
`,
		},
		{
			name:   "set-replicas-with-enum",
			setter: "replicas",
			value:  "baz",
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
          setBy: "package-default"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "foo"
          enumValues:
            foo: bar
            baz: biz
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
          setBy: "package-default"
 `,
			expected: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
          setBy: "package-default"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "baz"
          enumValues:
            foo: bar
            baz: biz
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
          setBy: "package-default"
`,
		},
		{
			name:   "set-replicas-fail",
			setter: "replicas",
			value:  "hello",
			isSet:  true,
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
          setBy: "package-default"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "foo"
          enumValues:
            foo: bar
            baz: biz
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
          setBy: "package-default"
 `,
			err: "hello does not match the possible values for replicas: [foo,baz]",
		},
		{
			name:   "error",
			setter: "replicas",
			err:    `setter "replicas" is not found`,
			isSet:  true,
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
 `,
			expected: `
openAPI:
  definitions:
    io.k8s.cli.setters.no-match-1':
      x-k8s-cli:
        setter:
          name: no-match-1
          value: "1"
    io.k8s.cli.setters.no-match-2':
      x-k8s-cli:
        setter:
          name: no-match-2
          value: "2"
 `,
		},

		{
			name:   "set-args-list",
			setter: "args",
			value:  "2",
			values: []string{"3", "4"},
			input: `
openAPI:
  definitions:
    io.k8s.cli.setters.args:
      type: array
      x-k8s-cli:
        setter:
          name: args
          listValues: ["1"]
          required: true
 `,
			expected: `
openAPI:
  definitions:
    io.k8s.cli.setters.args:
      type: array
      x-k8s-cli:
        setter:
          name: args
          listValues: ["2", "3", "4"]
          required: true
`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			in, err := yaml.Parse(test.input)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// invoke the setter
			instance := &SetOpenAPI{
				Name: test.setter, Value: test.value, ListValues: test.values,
				SetBy: test.setBy, Description: test.description, IsSet: test.isSet}
			result, err := instance.Filter(in)
			if test.err != "" {
				if !assert.EqualError(t, err, test.err) {
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

func TestValidateAgainstSchema(t *testing.T) {
	maxLength := int64(3)

	testCases := []struct {
		name             string
		setter           *setter
		schema           spec.SchemaProps
		shouldValidate   bool
		expectedErrorMsg string
	}{
		{
			name: "no schema",
			setter: &setter{
				Name:  "foo",
				Value: "bar",
			},
			schema:         spec.SchemaProps{},
			shouldValidate: true,
		},
		{
			name: "simple string value",
			setter: &setter{
				Name:  "foo",
				Value: "bar",
			},
			schema: spec.SchemaProps{
				Type: []string{"string"},
			},
			shouldValidate: true,
		},
		{
			name: "simple bool value",
			setter: &setter{
				Name:  "foo",
				Value: "false",
			},
			schema: spec.SchemaProps{
				Type: []string{"boolean"},
			},
			shouldValidate: true,
		},
		{
			name: "simple null value",
			setter: &setter{
				Name:  "foo",
				Value: "null",
			},
			schema: spec.SchemaProps{
				Type: []string{"null"},
			},
			shouldValidate: true,
		},
		{
			name: "bool value in yaml but not openapi",
			setter: &setter{
				Name:  "foo",
				Value: "yes",
			},
			schema: spec.SchemaProps{
				Type: []string{"string"},
			},
			shouldValidate: true,
		},
		{
			name: "number value should be accepted as integer",
			setter: &setter{
				Name:  "foo",
				Value: "45",
			},
			schema: spec.SchemaProps{
				Type: []string{"integer"},
			},
			shouldValidate: true,
		},
		{
			name: "string type allows string-specific validations",
			setter: &setter{
				Name:  "foo",
				Value: "1234",
			},
			schema: spec.SchemaProps{
				Type:      []string{"string"},
				MaxLength: &maxLength,
			},
			shouldValidate:   false,
			expectedErrorMsg: "foo in body should be at most 3 chars long",
		},
		{
			name: "list with int values",
			setter: &setter{
				Name:       "foo",
				ListValues: []string{"123", "456"},
			},
			schema: spec.SchemaProps{
				Type: []string{"array"},
				Items: &spec.SchemaOrArray{
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"integer"},
						},
					},
				},
			},
			shouldValidate: true,
		},
		{
			name: "list expecting int values, but with a string",
			setter: &setter{
				Name:       "foo",
				ListValues: []string{"123", "456", "abc"},
			},
			schema: spec.SchemaProps{
				Type: []string{"array"},
				Items: &spec.SchemaOrArray{
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"integer"},
						},
					},
				},
			},
			shouldValidate:   false,
			expectedErrorMsg: "foo in body must be of type integer",
		},
		{
			name: "all values should satisfy type string",
			setter: &setter{
				Name:  "foo",
				Value: "1234",
			},
			schema: spec.SchemaProps{
				Type: []string{"string"},
			},
			shouldValidate: true,
		},
		{
			name: "all values should satisfy type string even in arrays",
			setter: &setter{
				Name:       "foo",
				ListValues: []string{"123", "456"},
			},
			schema: spec.SchemaProps{
				Type: []string{"array"},
				Items: &spec.SchemaOrArray{
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"string"},
						},
					},
				},
			},
			shouldValidate: true,
		},
		{
			name: "List values without any schema",
			setter: &setter{
				Name:       "foo",
				ListValues: []string{"123", "true", "abc"},
			},
			schema:         spec.SchemaProps{},
			shouldValidate: true,
		},
	}

	for i := range testCases {
		test := testCases[i]
		t.Run(test.name, func(t *testing.T) {
			ext := &CliExtension{
				Setter: test.setter,
			}

			schema := &spec.Schema{
				SchemaProps: test.schema,
			}

			err := validateAgainstSchema(ext, schema)

			if test.shouldValidate {
				assert.NoError(t, err)
				return
			}

			if !assert.Error(t, err) {
				t.FailNow()
			}
			assert.Contains(t, err.Error(), test.expectedErrorMsg)
		})
	}
}
