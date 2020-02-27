// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"encoding/json"
	"strings"
	"testing"

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
			description: "if a type is specified for a setter, ensure the field is properly quoted",
			setter:      "foo",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.foo:
      x-k8s-cli:
        setter:
          name: foo
          value: "4"
      type: string
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
    foo: "4" # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
		},
		{
			name:        "set-foo-type-wrong",
			description: "if a type is specified for a setter, for the field to the type",
			setter:      "foo",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.foo:
      x-k8s-cli:
        setter:
          name: foo
          value: "4"
      type: boolean
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
    foo: !!bool 4 # {"$ref": "#/definitions/io.k8s.cli.setters.foo"}
 `,
		},
		{
			name:        "set-foo-no-type",
			description: "if a type is not specified for a setter, keep the existing quoting",
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
	_, err = openapi.AddSchema(j)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func TestSetOpenAPI_Filter(t *testing.T) {
	var tests = []struct {
		name        string
		setter      string
		value       string
		input       string
		expected    string
		description string
		setBy       string
		err         string
	}{
		{
			name:   "set-replicas",
			setter: "replicas",
			value:  "3",
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
`,
		},
		{
			name:        "set-replicas-description",
			setter:      "replicas",
			value:       "3",
			description: "hello world",
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
			err:    "no setter replicas found",
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
				Name: test.setter, Value: test.value,
				SetBy: test.setBy, Description: test.description}
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
