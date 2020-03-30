// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package prefixsuffix_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/prefixsuffix"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

var tests = []TestCase{
	{
		name: "prefix",
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
		filter: prefixsuffix.Filter{Prefix: "foo-"},
	},

	{
		name: "suffix",
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
		filter: prefixsuffix.Filter{Suffix: "-foo"},
	},

	{
		name: "prefix-suffix",
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
  name: bar-instance-foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: bar-instance-foo
`,
		filter: prefixsuffix.Filter{Prefix: "bar-", Suffix: "-foo"},
	},

	{
		name: "data-fieldspecs",
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
  name: foo-instance
a:
  b:
    c: foo-d
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: foo-instance
a:
  b:
    c: foo-d
`,
		filter: prefixsuffix.Filter{Prefix: "foo-"},
		fsslice: []types.FieldSpec{
			{
				Path: "a/b/c",
			},
		},
	},
}

type TestCase struct {
	name     string
	input    string
	expected string
	filter   prefixsuffix.Filter
	fsslice  types.FsSlice
}

var config = builtinconfig.MakeDefaultConfig()

func TestFilter(t *testing.T) {
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			test.filter.FsSlice = append(config.NamePrefix, test.fsslice...)
			if !assert.Equal(t,
				strings.TrimSpace(test.expected),
				strings.TrimSpace(
					filtertest_test.RunFilter(t, test.input, test.filter))) {
				t.FailNow()
			}
		})
	}
}
