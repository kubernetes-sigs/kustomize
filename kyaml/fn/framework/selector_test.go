// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestSelector(t *testing.T) {
	type Test struct {
		// Name is the name of the test
		Name string

		// Filter configures the selector
		Fn func(*framework.Selector)

		// ValueFoo is the value to substitute to select the foo resource
		ValueFoo string

		// ValueBar is the value to substitute to select the bar resource
		ValueBar string

		// Value is set by the test to either ValueFoo or ValueBar
		// and substituted into the selector
		Value string
	}
	tests := []Test{
		// Test the name template
		{
			Name: "names",
			Fn: func(s *framework.Selector) {
				s.Names = []string{"{{ .Value }}"}
			},
			ValueFoo: "foo",
			ValueBar: "bar",
		},

		// Test the kind template
		{
			Name: "kinds",
			Fn: func(s *framework.Selector) {
				s.Kinds = []string{"{{ .Value }}"}
			},
			ValueFoo: "StatefulSet",
			ValueBar: "Deployment",
		},

		// Test the apiVersion template
		{
			Name: "apiVersion",
			Fn: func(s *framework.Selector) {
				s.APIVersions = []string{"{{ .Value }}"}
			},
			ValueFoo: "apps/v1beta1",
			ValueBar: "apps/v1",
		},

		// Test the namespace template
		{
			Name: "namespaces",
			Fn: func(s *framework.Selector) {
				s.Namespaces = []string{"{{ .Value }}"}
			},
			ValueFoo: "foo-default",
			ValueBar: "bar-default",
		},

		// Test the annotations template
		{
			Name: "annotations",
			Fn: func(s *framework.Selector) {
				s.Annotations = map[string]string{"key": "{{ .Value }}"}
			},
			ValueFoo: "foo-a",
			ValueBar: "bar-a",
		},

		// Test the labels template
		{
			Name: "labels",
			Fn: func(s *framework.Selector) {
				s.Labels = map[string]string{"key": "{{ .Value }}"}
			},
			ValueFoo: "foo-l",
			ValueBar: "bar-l",
		},
	}

	// input is the input resources that are selected
	input := `
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
 name: foo
 namespace: foo-default
 annotations:
   key: foo-a
 labels:
   key: foo-l
---
apiVersion: apps/v1
kind: Deployment
metadata:
 name: bar
 namespace: bar-default
 annotations:
   key: bar-a
 labels:
   key: bar-l
`
	// expectedFoo is the expected output when the FooValue is substituted
	expectedFoo := `
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: foo
  namespace: foo-default
  annotations:
    key: foo-a
    config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/index: '0'
  labels:
    key: foo-l
`
	// expectedFoo is the expected output when the BarValue is substituted
	expectedBar := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  namespace: bar-default
  annotations:
    key: bar-a
    config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/index: '1'
  labels:
    key: bar-l
`

	// Run the tests by substituting the FooValues
	var err error
	for i := range tests {
		test := tests[i]
		t.Run(tests[i].Name+"-foo", func(t *testing.T) {
			test.Value = test.ValueFoo
			var out bytes.Buffer
			rw := &kio.ByteReadWriter{
				Reader:                bytes.NewBufferString(input),
				Writer:                &out,
				KeepReaderAnnotations: true,
			}
			p := func(rl *framework.ResourceList) error {
				s := &framework.Selector{TemplateData: test}
				test.Fn(s)
				rl.Items, err = s.Filter(rl.Items)
				return err
			}

			require.NoError(t, framework.Execute(framework.ResourceListProcessorFunc(p), rw))
			require.Equal(t, strings.TrimSpace(expectedFoo), strings.TrimSpace(out.String()))
		})
	}

	// Run the tests by substituting the BarValues
	for i := range tests {
		test := tests[i]
		t.Run(tests[i].Name+"-bar", func(t *testing.T) {
			test.Value = test.ValueBar
			var out bytes.Buffer
			rw := &kio.ByteReadWriter{
				Reader:                bytes.NewBufferString(input),
				Writer:                &out,
				KeepReaderAnnotations: true,
			}

			p := func(rl *framework.ResourceList) error {
				s := &framework.Selector{TemplateData: test}
				test.Fn(s)
				rl.Items, err = s.Filter(rl.Items)
				return err
			}
			require.NoError(t, framework.Execute(framework.ResourceListProcessorFunc(p), rw))
			require.Equal(t, strings.TrimSpace(expectedBar), strings.TrimSpace(out.String()))
		})
	}
}

func TestAndOrSelector_Composition(t *testing.T) {
	// This selector should pick the "prime-target" deployment by name
	// as well as any resources with the given labels or annotations regardless of kind
	s := framework.MatchAny(
		framework.MatchAll(
			framework.GVKMatcher("apps/v1/Deployment"),
			framework.NameMatcher("prime-target"),
		),
		framework.MatchAny(
			framework.LabelMatcher(map[string]string{
				"select": "yes",
			}),
			framework.AnnotationMatcher(map[string]string{
				"example.io/select": "yes",
			}),
		),
	)

	input, err := kio.FromBytes([]byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prime-target
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: exclude-one
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: exclude-two
  labels:
    select: no
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: extra-target
  labels:
    select: yes
---
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: prime-target
data:
  shouldSelect: false
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extra-target-one
  labels:
    select: yes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extra-target-two
  annotations:
    example.io/select: yes
`))
	require.NoError(t, err)
	result, err := s.Filter(input)
	require.NoError(t, err)

	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prime-target
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: extra-target
  labels:
    select: yes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extra-target-one
  labels:
    select: yes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extra-target-two
  annotations:
    example.io/select: yes
`
	resultStr, err := kio.StringAll(result)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(resultStr))
}

func TestAndOrSelector_CompositionTemplated(t *testing.T) {
	// This selector should pick the "prime-target" deployment by name
	// as well as any resources with the given labels or annotations regardless of kind
	// Note: very similar to above test, but uses verbose expression to access templating
	type templateStruct struct {
		GVK             string
		Name            string
		LabelValue      string
		AnnotationValue string
	}

	s := framework.OrSelector{
		// This should get propagated to matchers without explicit data
		TemplateData: &templateStruct{
			GVK:             "apps/v1/Oops",
			Name:            "extra-target",
			LabelValue:      "yes",
			AnnotationValue: "yes",
		},
		Matchers: []framework.ResourceMatcher{
			&framework.AndSelector{
				TemplateData: &templateStruct{
					GVK:  "apps/v1/Deployment",
					Name: "prime-target",
				},
				Matchers: []framework.ResourceMatcher{
					framework.GVKMatcher("{{.GVK}}"),
					framework.NameMatcher("{{.Name}}"),
				},
			},
			&framework.OrSelector{
				Matchers: []framework.ResourceMatcher{
					framework.LabelMatcher(map[string]string{
						"select": "{{.LabelValue}}",
					}),
					framework.AnnotationMatcher(map[string]string{
						"example.io/select": "{{.AnnotationValue}}",
					}),
				},
			},
		},
	}

	input, err := kio.FromBytes([]byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prime-target
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: exclude-one
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: exclude-two
  labels:
    select: no
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: extra-target
  labels:
    select: yes
---
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: prime-target
data:
  shouldSelect: false
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extra-target-one
  labels:
    select: yes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extra-target-two
  annotations:
    example.io/select: yes
`))
	require.NoError(t, err)
	result, err := s.Filter(input)
	require.NoError(t, err)

	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prime-target
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: extra-target
  labels:
    select: yes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extra-target-one
  labels:
    select: yes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extra-target-two
  annotations:
    example.io/select: yes
`
	resultStr, err := kio.StringAll(result)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(resultStr))
}

func TestMatchersAsFilters(t *testing.T) {
	input, err := kio.FromBytes([]byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: target
  labels:
    select: me
---
apiVersion: extensions/v1beta2
kind: Deployment
metadata:
  name: exclude
  labels:
    select: no
`))
	require.NoError(t, err)

	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: target
  labels:
    select: me
`
	matchers := map[string]framework.ResourceMatcher{
		"slice": framework.NameMatcher("target"),
		"map":   framework.LabelMatcher(map[string]string{"select": "me"}),
		"func": framework.ResourceMatcherFunc(func(node *yaml.RNode) bool {
			v := node.Field("apiVersion").Value
			return strings.TrimSpace(v.MustString()) == "apps/v1"
		}),
	}
	for desc, m := range matchers {
		matcher := m
		t.Run(desc, func(t *testing.T) {
			result, err := matcher.Filter(input)
			require.NoError(t, err)
			resultStr, err := kio.StringAll(result)
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(resultStr))
		})
	}
}
