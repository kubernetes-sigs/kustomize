// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package annotations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var annosFs = builtinconfig.MakeDefaultConfig().CommonAnnotations

func TestAnnotations_Filter(t *testing.T) {
	mutationTrackStub := filtertest_test.MutationTrackerStub{}
	testCases := map[string]struct {
		input                string
		expectedOutput       string
		filter               Filter
		fsslice              types.FsSlice
		setEntryCallback     func(key, value, tag string, node *yaml.RNode)
		expectedSetEntryArgs []filtertest_test.SetValueArg
	}{
		"add": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
    auto: ford
    bean: cannellini
    clown: emmett kelley
    dragon: smaug
`,
			filter: Filter{Annotations: annoMap{
				"clown":  "emmett kelley",
				"auto":   "ford",
				"dragon": "smaug",
				"bean":   "cannellini",
			}},
		},
		"update": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: superman
    fiend: luthor
    bean: cannellini
    clown: emmett kelley
`,
			filter: Filter{Annotations: annoMap{
				"clown": "emmett kelley",
				"hero":  "superman",
				"fiend": "luthor",
				"bean":  "cannellini",
			}},
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
  annotations:
    sleater: kinney
a:
  b:
    sleater: kinney
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  annotations:
    sleater: kinney
a:
  b:
    sleater: kinney
`,
			filter: Filter{Annotations: annoMap{
				"sleater": "kinney",
			}},
			fsslice: []types.FieldSpec{
				{
					Path:               "a/b",
					CreateIfNotPresent: true,
				},
			},
		},

		"number": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
    2: ford
    clown: "1"
`,
			filter: Filter{Annotations: annoMap{
				"clown": "1",
				"2":     "ford",
			}},
		},

		// test quoting of values which are not considered strings in yaml 1.1
		"yaml_1_1_compatibility": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
    a: "y"
    b: y1
    c: "yes"
    d: yes1
    e: "true"
    f: true1
`,
			filter: Filter{Annotations: annoMap{
				"a": "y",
				"b": "y1",
				"c": "yes",
				"d": "yes1",
				"e": "true",
				"f": "true1",
			}},
		},

		// test quoting of values which are not considered strings in yaml 1.1
		"null_annotations": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations: null
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    a: a1
    b: b1
`,
			filter: Filter{Annotations: annoMap{
				"a": "a1",
				"b": "b1",
			}},
		},

		// test usage of SetEntryCallback
		"set_entry_callback": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    a: a1
    b: b1
spec:
  template:
    metadata:
      annotations:
        a: a1
        b: b1
`,
			filter: Filter{
				Annotations: annoMap{
					"a": "a1",
					"b": "b1",
				},
			},
			setEntryCallback: mutationTrackStub.MutationTracker,
			fsslice: []types.FieldSpec{
				{
					Path:               "spec/template/metadata/annotations",
					CreateIfNotPresent: true,
				},
			},
			expectedSetEntryArgs: []filtertest_test.SetValueArg{
				{
					Key:      "a",
					Value:    "a1",
					Tag:      "!!str",
					NodePath: []string{"metadata", "annotations"},
				},
				{
					Key:      "a",
					Value:    "a1",
					Tag:      "!!str",
					NodePath: []string{"spec", "template", "metadata", "annotations"},
				},
				{
					Key:      "b",
					Value:    "b1",
					Tag:      "!!str",
					NodePath: []string{"metadata", "annotations"},
				},
				{
					Key:      "b",
					Value:    "b1",
					Tag:      "!!str",
					NodePath: []string{"spec", "template", "metadata", "annotations"},
				},
			},
		},
	}

	for tn, tc := range testCases {
		mutationTrackStub.Reset()
		t.Run(tn, func(t *testing.T) {
			filter := tc.filter
			filter.WithMutationTracker(tc.setEntryCallback)
			filter.FsSlice = append(annosFs, tc.fsslice...) //nolint:gocritic
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput),
				strings.TrimSpace(filtertest_test.RunFilter(t, tc.input, filter))) {
				t.FailNow()
			}
			if !assert.Equal(t, tc.expectedSetEntryArgs, mutationTrackStub.SetValueArgs()) {
				t.FailNow()
			}
		})
	}
}
