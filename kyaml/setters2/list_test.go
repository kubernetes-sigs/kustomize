// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func TestList(t *testing.T) {
	var tests = []struct {
		name     string
		setter   string
		openapi  string
		input    string
		expected []SetterDefinition
	}{
		{
			name: "list-replicas",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
      description: "hello world"
 `,
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
 `,
			expected: []SetterDefinition{
				{Name: "replicas", Value: "3", SetBy: "me", Description: "hello world", Count: 1},
			},
		},
		{
			name: "list-multiple",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: "hello world 1"
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me1
    io.k8s.cli.setters.image:
      description: "hello world 2"
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
          setBy: me2
    io.k8s.cli.setters.tag:
      description: "hello world 3"
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
          setBy: me3
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
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx # {"$ref": "#/definitions/io.k8s.cli.setters.image"}
 `,
			expected: []SetterDefinition{
				{Name: "image", Value: "nginx", SetBy: "me2", Description: "hello world 2", Count: 2},
				{Name: "replicas", Value: "3", SetBy: "me1", Description: "hello world 1", Count: 1},
				{Name: "tag", Value: "1.7.9", SetBy: "me3", Description: "hello world 3", Count: 1},
			},
		},
		{
			name: "list-multiple-resources",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: "hello world 1"
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me1
    io.k8s.cli.setters.image:
      description: "hello world 2"
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
          setBy: me2
    io.k8s.cli.setters.tag:
      description: "hello world 3"
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
          setBy: me3
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
  name: nginx-deployment-1
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx # {"$ref": "#/definitions/io.k8s.cli.setters.image"}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-2
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx
`,
			expected: []SetterDefinition{
				{Name: "image", Value: "nginx", SetBy: "me2", Description: "hello world 2", Count: 3},
				{Name: "replicas", Value: "3", SetBy: "me1", Description: "hello world 1", Count: 2},
				{Name: "tag", Value: "1.7.9", SetBy: "me3", Description: "hello world 3", Count: 2},
			},
		},
		{
			name: "list-name",
			openapi: `
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: "hello world 1"
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me1
    io.k8s.cli.setters.image:
      description: "hello world 2"
      x-k8s-cli:
        setter:
          name: image
          value: "nginx"
          setBy: me2
    io.k8s.cli.setters.tag:
      description: "hello world 3"
      x-k8s-cli:
        setter:
          name: tag
          value: "1.7.9"
          setBy: me3
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
  name: nginx-deployment-1
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx # {"$ref": "#/definitions/io.k8s.cli.setters.image"}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-2
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
      - name: nginx2
        image: nginx
`,
			setter: "image",
			expected: []SetterDefinition{
				{Name: "image", Value: "nginx", SetBy: "me2", Description: "hello world 2", Count: 3},
			},
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			defer openapi.ResetOpenAPI()
			initSchema(t, test.openapi)

			f, err := ioutil.TempFile("", "k8s-cli-")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.Remove(f.Name())
			err = ioutil.WriteFile(f.Name(), []byte(test.openapi), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
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

			// invoke the setter
			instance := &List{Name: test.setter}
			err = instance.List(f.Name(), r.Name())
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t, test.expected, instance.Setters) {
				t.FailNow()
			}
		})
	}
}
