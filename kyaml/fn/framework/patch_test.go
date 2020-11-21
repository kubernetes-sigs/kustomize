// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
	"sigs.k8s.io/kustomize/kyaml/testutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestPatchTemplate(t *testing.T) {
	// TODO: make this test pass on windows -- current failure seems spurious
	testutil.SkipWindows(t)

	cmdFn := func() *cobra.Command {
		type api struct {
			Selector framework.Selector `json:"selector" yaml:"selector"`
			A        string             `json:"a" yaml:"a"`
			B        string             `json:"b" yaml:"b"`
			Special  string             `json:"special" yaml:"special"`
			LongList bool
		}
		var config api
		filter := framework.Selector{
			// this is a special manual filter for the Selector for when the built-in matchers
			// are insufficient
			Filter: func(rn *yaml.RNode) bool {
				m, _ := rn.GetMeta()
				return config.Special != "" && m.Annotations["foo"] == config.Special
			},
		}
		return framework.TemplateCommand{
			API: &config,
			PreProcess: func(rl *framework.ResourceList) error {
				// do some extra processing based on the inputs
				config.LongList = len(rl.Items) > 2
				return nil
			},
			PatchTemplates: []framework.PatchTemplate{
				{
					// Apply these rendered patches
					Template: template.Must(template.New("test").Parse(`
spec:
  template:
    spec:
      containers:
      - name: foo
        image: example/sidecar:{{ .B }}
---
metadata:
  annotations:
    patched: '{{ .A }}'
{{- if .LongList }}
    long: 'true'
{{- end }}
`)),
					// Use the selector from the input
					Selector: &config.Selector,
				},
				{
					// Apply these rendered patches
					Template: template.Must(template.New("test").Parse(`
metadata:
  annotations:
    filterPatched: '{{ .A }}'
`)),
					// Use an explicit selector
					Selector: &filter,
				},
			},
		}.GetCommand()
	}

	frameworktestutil.ResultsChecker{Command: cmdFn, TestDataDirectory: "patchtestdata"}.Assert(t)
}

func TestSelector(t *testing.T) {
	type Test struct {
		// Name is the name of the test
		Name string

		// Fn configures the selector
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
			rl := &framework.ResourceList{
				FunctionConfig: test,
				Reader:         bytes.NewBufferString(input),
				Writer:         &out,
			}
			if !assert.NoError(t, rl.Read()) {
				t.FailNow()
			}
			s := &framework.Selector{TemplatizeValues: true}
			test.Fn(s)
			rl.Items, err = s.GetMatches(rl)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.NoError(t, rl.Write()) {
				t.FailNow()
			}
			if !assert.Equal(t, strings.TrimSpace(expectedFoo), strings.TrimSpace(out.String())) {
				t.FailNow()
			}
		})
	}

	// Run the tests by substituting the BarValues
	for i := range tests {
		test := tests[i]
		t.Run(tests[i].Name+"-bar", func(t *testing.T) {
			test.Value = test.ValueBar
			var out bytes.Buffer
			rl := &framework.ResourceList{
				FunctionConfig: test,
				Reader:         bytes.NewBufferString(input),
				Writer:         &out,
			}
			if !assert.NoError(t, rl.Read()) {
				t.FailNow()
			}
			s := &framework.Selector{TemplatizeValues: true}
			test.Fn(s)
			rl.Items, err = s.GetMatches(rl)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.NoError(t, rl.Write()) {
				t.FailNow()
			}
			if !assert.Equal(t, strings.TrimSpace(expectedBar), strings.TrimSpace(out.String())) {
				t.FailNow()
			}
		})
	}
}
