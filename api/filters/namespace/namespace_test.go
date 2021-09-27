// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package namespace_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/namespace"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

var tests = []TestCase{
	{
		name: "add",
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
  name: instance
  namespace: foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  namespace: foo
`,
		filter: namespace.Filter{Namespace: "foo"},
	},

	{
		name: "null_ns",
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  namespace: null
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  namespace: null
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  namespace: foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  namespace: foo
`,
		filter: namespace.Filter{Namespace: "foo"},
	},

	{
		name: "add-recurse",
		input: `
apiVersion: example.com/v1
kind: Foo
---
apiVersion: example.com/v1
kind: Bar
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  namespace: foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  namespace: foo
`,
		filter: namespace.Filter{Namespace: "foo"},
	},

	{
		name: "update",
		input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  # update this namespace
  namespace: bar
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  namespace: bar
`,
		expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  # update this namespace
  namespace: foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  namespace: foo
`,
		filter: namespace.Filter{Namespace: "foo"},
	},

	{
		name: "update-rolebinding",
		input: `
apiVersion: example.com/v1
kind: RoleBinding
subjects:
- name: default
---
apiVersion: example.com/v1
kind: RoleBinding
subjects:
- name: default
  namespace: foo
---
apiVersion: example.com/v1
kind: RoleBinding
subjects:
- name: something
---
apiVersion: example.com/v1
kind: RoleBinding
subjects:
- name: something
  namespace: foo
`,
		expected: `
apiVersion: example.com/v1
kind: RoleBinding
subjects:
- name: default
  namespace: bar
metadata:
  namespace: bar
---
apiVersion: example.com/v1
kind: RoleBinding
subjects:
- name: default
  namespace: bar
metadata:
  namespace: bar
---
apiVersion: example.com/v1
kind: RoleBinding
subjects:
- name: something
metadata:
  namespace: bar
---
apiVersion: example.com/v1
kind: RoleBinding
subjects:
- name: something
  namespace: foo
metadata:
  namespace: bar
`,
		filter: namespace.Filter{Namespace: "bar"},
	},

	{
		name: "update-clusterrolebinding",
		input: `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
subjects:
- name: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
subjects:
- name: default
  namespace: foo
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
subjects:
- name: something
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
subjects:
- name: something
  namespace: foo
`,
		expected: `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
subjects:
- name: default
  namespace: bar
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
subjects:
- name: default
  namespace: bar
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
subjects:
- name: something
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
subjects:
- name: something
  namespace: foo
`,
		filter: namespace.Filter{Namespace: "bar"},
	},

	{
		name: "data-fieldspecs",
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
  name: instance
  namespace: foo
a:
  b:
    c: foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  namespace: foo
a:
  b:
    c: foo
`,
		filter: namespace.Filter{Namespace: "foo"},
		fsslice: []types.FieldSpec{
			{
				Path:               "a/b/c",
				CreateIfNotPresent: true,
			},
		},
	},
}

type TestCase struct {
	name     string
	input    string
	expected string
	filter   namespace.Filter
	fsslice  types.FsSlice
}

var config = builtinconfig.MakeDefaultConfig()

func TestNamespace_Filter(t *testing.T) {
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			test.filter.FsSlice = append(config.NameSpace, test.fsslice...)
			if !assert.Equal(t,
				strings.TrimSpace(test.expected),
				strings.TrimSpace(
					filtertest_test.RunFilter(t, test.input, test.filter))) {
				t.FailNow()
			}
		})
	}
}
