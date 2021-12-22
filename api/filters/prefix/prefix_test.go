// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package prefix_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/prefix"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

var tests = map[string]TestCase{
	"prefix": {
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: foo-instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: foo-instance
`,
		filter: prefix.Filter{
			Prefix:    "foo-",
			FieldSpec: types.FieldSpec{Path: "metadata/name"},
		},
	},

	"data-fieldspecs": {
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
a:
  b:
    c: d
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
a:
  b:
    c: d
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
a:
  b:
    c: foo-d
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
a:
  b:
    c: foo-d
`,
		filter: prefix.Filter{
			Prefix:    "foo-",
			FieldSpec: types.FieldSpec{Path: "a/b/c"},
		},
	},
}

type TestCase struct {
	input    string
	expected string
	filter   prefix.Filter
}

func TestFilter(t *testing.T) {
	for name := range tests {
		test := tests[name]
		t.Run(name, func(t *testing.T) {
			if !assert.Equal(t,
				strings.TrimSpace(test.expected),
				strings.TrimSpace(
					filtertest_test.RunFilter(t, test.input, test.filter))) {
				t.FailNow()
			}
		})
	}
}
