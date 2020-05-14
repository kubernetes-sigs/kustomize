// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that non-UTF8 characters in comments don't cause failures
func TestRNode_GetMeta_UTF16(t *testing.T) {
	sr, err := Parse(`apiVersion: rbac.istio.io/v1alpha1
kind: ServiceRole
metadata:
  name: wildcard
  namespace: default
  # If set to [“*”], it refers to all services in the namespace
  annotations:
    foo: bar
spec:
  rules:
    # There is one service in default namespace, should not result in a validation error
    # If set to [“*”], it refers to all services in the namespace
    - services: ["*"]
      methods: ["GET", "HEAD"]
`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	actual, err := sr.GetMeta()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := ResourceMeta{
		APIVersion: "rbac.istio.io/v1alpha1",
		Kind:       "ServiceRole",
		ObjectMeta: ObjectMeta{
			Name:        "wildcard",
			Namespace:   "default",
			Annotations: map[string]string{"foo": "bar"},
		},
	}
	if !assert.Equal(t, expected, actual) {
		t.FailNow()
	}
}

func TestRNode_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		output   string
	}{
		{
			testName: "simple document",
			input:    `{"hello":"world"}`,
			output:   `hello: world`,
		},
		{
			testName: "nested structure",
			input: `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "my-deployment",
    "namespace": "default"
  }
}
`,
			output: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: default
`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			instance := &RNode{}
			err := instance.UnmarshalJSON([]byte(tc.input))
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actual, err := instance.String()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t,
				strings.TrimSpace(tc.output), strings.TrimSpace(actual)) {
				t.FailNow()
			}
		})
	}
}

func TestRNode_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		ydoc string
		want string
	}{
		{
			name: "object",
			ydoc: `
hello: world
`,
			want: `{"hello":"world"}`,
		},
		{
			name: "array",
			ydoc: `
- name: s1
- name: s2
`,
			want: `[{"name":"s1"},{"name":"s2"}]`,
		},
	}
	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			instance, err := Parse(tt.ydoc)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actual, err := instance.MarshalJSON()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t,
				strings.TrimSpace(tt.want), strings.TrimSpace(string(actual))) {
				t.FailNow()
			}
		})
	}
}
