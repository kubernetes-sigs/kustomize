// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package suffix_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/suffix"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

var tests = map[string]TestCase{
	"suffix": {
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
  name: instance-foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance-foo
`,
		filter: suffix.Filter{
			Suffix:    "-foo",
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
    c: d-foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
a:
  b:
    c: d-foo
`,
		filter: suffix.Filter{
			Suffix:    "-foo",
			FieldSpec: types.FieldSpec{Path: "a/b/c"},
		},
	},
}

type TestCase struct {
	input    string
	expected string
	filter   suffix.Filter
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
