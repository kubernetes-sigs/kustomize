// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
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
