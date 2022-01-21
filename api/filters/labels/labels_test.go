// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package labels

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type setEntryArg struct {
	Key      string
	Value    string
	Tag      string
	NodePath []string
}

var setEntryArgs []setEntryArg

func setEntryCallbackStub(key, value, tag string, node *yaml.RNode) {
	setEntryArgs = append(setEntryArgs, setEntryArg{
		Key:      key,
		Value:    value,
		Tag:      tag,
		NodePath: node.FieldPath(),
	})
}

func TestLabels_Filter(t *testing.T) {
	testCases := map[string]struct {
		input                string
		expectedOutput       string
		filter               Filter
		setEntryCallback     func(key, value, tag string, node *yaml.RNode)
		expectedSetEntryArgs []setEntryArg
	}{
		"add": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    hero: batman
    fiend: riddler
    auto: ford
    bean: cannellini
    clown: emmett kelley
    dragon: smaug
`,
			filter: Filter{
				Labels: labelMap{
					"clown":  "emmett kelley",
					"auto":   "ford",
					"dragon": "smaug",
					"bean":   "cannellini",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
				},
			},
		},
		"update": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    hero: superman
    fiend: luthor
    bean: cannellini
    clown: emmett kelley
`,
			filter: Filter{
				Labels: labelMap{
					"clown": "emmett kelley",
					"hero":  "superman",
					"fiend": "luthor",
					"bean":  "cannellini",
				}, FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
				},
			},
		},

		"data-fieldspecs": {
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
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    sleater: kinney
a:
  b:
    sleater: kinney
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  labels:
    sleater: kinney
a:
  b:
    sleater: kinney
`,
			filter: Filter{
				Labels: labelMap{
					"sleater": "kinney",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
					{
						Path:               "a/b",
						CreateIfNotPresent: true,
					},
				},
			},
		},

		"fieldSpecWithKind": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v2
kind: Bar
metadata:
  name: instance
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    cheese: cheddar
---
apiVersion: example.com/v2
kind: Bar
metadata:
  name: instance
  labels:
    cheese: cheddar
a:
  b:
    cheese: cheddar
`,
			filter: Filter{
				Labels: labelMap{
					"cheese": "cheddar",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
					{
						Gvk: resid.Gvk{
							Kind: "Bar",
						},
						Path:               "a/b",
						CreateIfNotPresent: true,
					},
				},
			},
		},

		"fieldSpecWithVersion": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v2
kind: Bar
metadata:
  name: instance
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    cheese: cheddar
a:
  b:
    cheese: cheddar
---
apiVersion: example.com/v2
kind: Bar
metadata:
  name: instance
  labels:
    cheese: cheddar
`,
			filter: Filter{
				Labels: labelMap{
					"cheese": "cheddar",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
					{
						Gvk: resid.Gvk{
							Version: "v1",
						},
						Path:               "a/b",
						CreateIfNotPresent: true,
					},
				},
			},
		},
		"fieldSpecWithVersionInConfigButNoGroupInData": {
			input: `
apiVersion: v1
kind: Foo
metadata:
  name: instance
---
apiVersion: v2
kind: Bar
metadata:
  name: instance
`,
			expectedOutput: `
apiVersion: v1
kind: Foo
metadata:
  name: instance
  labels:
    cheese: cheddar
a:
  b:
    cheese: cheddar
---
apiVersion: v2
kind: Bar
metadata:
  name: instance
  labels:
    cheese: cheddar
`,
			filter: Filter{
				Labels: labelMap{
					"cheese": "cheddar",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
					{
						Gvk: resid.Gvk{
							Version: "v1",
						},
						Path:               "a/b",
						CreateIfNotPresent: true,
					},
				},
			},
		},

		"number": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    hero: batman
    fiend: riddler
    1: emmett kelley
    auto: "2"
`,
			filter: Filter{
				Labels: labelMap{
					"1":    "emmett kelley",
					"auto": "2",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
				},
			},
		},

		// test quoting of values which are not considered strings in yaml 1.1
		"yaml_1_1_compatibility": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    hero: batman
    fiend: riddler
    a: "y"
    b: y1
    c: "yes"
    d: yes1
    e: "true"
    f: true1
`,
			filter: Filter{
				Labels: labelMap{
					"a": "y",
					"b": "y1",
					"c": "yes",
					"d": "yes1",
					"e": "true",
					"f": "true1",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
				},
			},
		},

		"null_labels": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels: null
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    a: a1
`,
			filter: Filter{
				Labels: labelMap{
					"a": "a1",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
				},
			},
		},

		// test usage of SetEntryCallback
		"set_entry_callback": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    witcher: geralt
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  labels:
    witcher: geralt
    mage: yennefer
a:
  b:
    mage: yennefer
`,
			filter: Filter{
				Labels: labelMap{
					"mage": "yennefer",
				},
				FsSlice: []types.FieldSpec{
					{
						Path:               "metadata/labels",
						CreateIfNotPresent: true,
					},
					{
						Path:               "a/b",
						CreateIfNotPresent: true,
					},
				},
			},
			setEntryCallback: setEntryCallbackStub,
			expectedSetEntryArgs: []setEntryArg{
				{
					Key:      "mage",
					Value:    "yennefer",
					Tag:      "!!str",
					NodePath: []string{"metadata", "labels"},
				},
				{
					Key:      "mage",
					Value:    "yennefer",
					Tag:      "!!str",
					NodePath: []string{"a", "b"},
				},
			},
		},
	}

	for tn, tc := range testCases {
		setEntryArgs = nil
		t.Run(tn, func(t *testing.T) {
			tc.filter.WithMutationTracker(tc.setEntryCallback)
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput),
				strings.TrimSpace(filtertest_test.RunFilter(t, tc.input, tc.filter))) {
				t.FailNow()
			}
			if !assert.Equal(t, tc.expectedSetEntryArgs, setEntryArgs) {
				t.FailNow()
			}
		})
	}
}
