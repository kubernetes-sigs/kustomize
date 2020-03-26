// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package annotations

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestAnnotations_Filter(t *testing.T) {
	testCases := map[string]struct {
		input          string
		expectedOutput string
		filter         Filter
		fsslice        types.FsSlice
	}{
		"add": {
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
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  annotations:
    sleater: kinney
`,
			filter: Filter{Annotations: map[string]string{
				"sleater": "kinney",
			}},
		},
		"update": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    foo: foo
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    foo: bar
`,
			filter: Filter{Annotations: map[string]string{
				"foo": "bar",
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
			filter: Filter{Annotations: map[string]string{
				"sleater": "kinney",
			}},
			fsslice: []types.FieldSpec{
				{
					Path:               "a/b",
					CreateIfNotPresent: true,
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			config := builtinconfig.MakeDefaultConfig()

			filter := tc.filter
			filter.FsSlice = append(config.CommonAnnotations, tc.fsslice...)

			var out bytes.Buffer
			rw := kio.ByteReadWriter{
				Reader: bytes.NewBufferString(tc.input),
				Writer: &out,
			}

			err := kio.Pipeline{
				Inputs:  []kio.Reader{&rw},
				Filters: []kio.Filter{filter},
				Outputs: []kio.Writer{&rw},
			}.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput),
				strings.TrimSpace(out.String())) {
				t.FailNow()
			}
		})
	}
}

func TestAnnotations_Filter_Multiple(t *testing.T) {
	input := `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
`
	annos := map[string]string{
		"sleater": "kinney",
		"sonic":   "youth",
	}
	config := builtinconfig.MakeDefaultConfig()
	filter := Filter{Annotations: annos}
	filter.FsSlice = config.CommonAnnotations

	var out bytes.Buffer
	rw := kio.ByteReadWriter{
		Reader: bytes.NewBufferString(input),
		Writer: &out,
	}

	err := kio.Pipeline{
		Inputs:  []kio.Reader{&rw},
		Filters: []kio.Filter{filter},
		Outputs: []kio.Writer{&rw},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assertHasAnnotation(t, out.String(), annos)
}

func assertHasAnnotation(t *testing.T, y string, exp map[string]string) bool {
	var out bytes.Buffer
	rw := kio.ByteReadWriter{
		Reader: bytes.NewBufferString(y),
		Writer: &out,
	}
	filter := &captureAnnotationFilter{
		annotations: make(map[string]annotations),
	}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&rw},
		Filters: []kio.Filter{filter},
		Outputs: []kio.Writer{&rw},
	}.Execute()
	if err != nil {
		t.Error(err)
		return false
	}

	for name, annos := range filter.annotations {
		for key, val := range exp {
			v, found := annos[key]
			if !found {
				t.Errorf("expected annotation with key %s in object %s, but didn't find it",
					key, name)
				return false
			}
			if want, got := val, v; got != want {
				t.Errorf("exected annotation %s in object %s to have value %s, but found %s",
					key, name, want, got)
				return false
			}
		}
	}
	return true
}

type annotations map[string]string

type captureAnnotationFilter struct {
	annotations map[string]annotations
}

func (c captureAnnotationFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode,
	error) {
	for _, n := range nodes {
		meta, err := n.GetMeta()
		if err != nil {
			return nodes, err
		}
		name := meta.Name
		annos := meta.Annotations
		c.annotations[name] = annos
	}
	return nodes, nil
}
