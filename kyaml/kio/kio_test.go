// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	. "sigs.k8s.io/kustomize/kyaml/kio"
)

func TestPipe(t *testing.T) {
	p := Pipeline{
		Inputs:  []Reader{},
		Filters: []Filter{},
		Outputs: []Writer{},
	}

	err := p.Execute()
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
}

type mockCallback struct {
	mock.Mock
}

func (c *mockCallback) Callback(op Filter) {
	c.Called(op)
}

func TestPipelineWithCallback(t *testing.T) {
	input := ResourceNodeSlice{yaml.MakeNullNode()}
	noopFilter1 := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		return nodes, nil
	}
	noopFilter2 := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		return nodes, nil
	}
	filters := []Filter{
		FilterFunc(noopFilter1),
		FilterFunc(noopFilter2),
	}
	p := Pipeline{
		Inputs:  []Reader{input},
		Filters: filters,
		Outputs: []Writer{},
	}

	callback := mockCallback{}
	// setup expectations. `Times` means the function is called no more than `times`.
	callback.On("Callback", mock.Anything).Times(len(filters))

	err := p.ExecuteWithCallback(callback.Callback)

	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	callback.AssertNumberOfCalls(t, "Callback", len(filters))

	// assert filters are called in the order they are defined.
	for i, filter := range filters {
		assert.Equal(
			t,
			reflect.ValueOf(callback.Calls[i].Arguments[0]).Pointer(),
			reflect.ValueOf(filter).Pointer(),
		)
	}
}

func TestEmptyInput(t *testing.T) {
	actual := &bytes.Buffer{}
	output := ByteWriter{
		Sort:               true,
		WrappingKind:       ResourceListKind,
		WrappingAPIVersion: ResourceListAPIVersion,
	}
	output.Writer = actual

	p := Pipeline{
		Outputs: []Writer{output},
	}

	err := p.Execute()
	if err != nil {
		t.Fatal(err)
	}

	expected := `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
`

	if !assert.Equal(t,
		strings.TrimSpace(expected), strings.TrimSpace(actual.String())) {
		t.FailNow()
	}
}

func TestEmptyInputWithFilter(t *testing.T) {
	actual := &bytes.Buffer{}
	output := ByteWriter{
		Sort:               true,
		WrappingKind:       ResourceListKind,
		WrappingAPIVersion: ResourceListAPIVersion,
	}
	output.Writer = actual

	filters := []Filter{
		FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			nodes = append(nodes, yaml.NewMapRNode(&map[string]string{
				"foo": "bar",
			}))
			return nodes, nil
		}),
		FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) { return nodes, nil }),
	}

	p := Pipeline{
		Outputs: []Writer{output},
		Filters: filters,
	}

	err := p.Execute()
	if err != nil {
		t.Fatal(err)
	}

	expected := `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- foo: bar
`

	if !assert.Equal(t,
		strings.TrimSpace(expected), strings.TrimSpace(actual.String())) {
		t.FailNow()
	}
}

func TestContinueOnEmptyBehavior(t *testing.T) {
	cases := map[string]struct {
		continueOnEmptyResult bool
		expected              string
	}{
		"quit on empty":     {continueOnEmptyResult: false, expected: ""},
		"continue on empty": {continueOnEmptyResult: true, expected: "foo: bar"},
	}
	for _, tc := range cases {
		actual := &bytes.Buffer{}
		output := ByteWriter{Writer: actual}

		generatorFunc := FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			nodes = append(nodes, yaml.NewMapRNode(&map[string]string{
				"foo": "bar",
			}))
			return nodes, nil
		})
		emptyFunc := FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) { return nodes, nil })

		p := Pipeline{
			Outputs:               []Writer{output},
			Filters:               []Filter{emptyFunc, generatorFunc},
			ContinueOnEmptyResult: tc.continueOnEmptyResult,
		}

		err := p.Execute()
		if err != nil {
			t.Fatal(err)
		}

		if !assert.Equal(t,
			tc.expected, strings.TrimSpace(actual.String())) {
			t.Fail()
		}
	}
}

func TestLegacyAnnotationReconciliation(t *testing.T) {
	noopFilter1 := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		return nodes, nil
	}
	noopFilter2 := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		return nodes, nil
	}
	changeInternalAnnos := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, rn := range nodes {
			if err := rn.PipeE(yaml.SetAnnotation(kioutil.PathAnnotation, "new")); err != nil {
				return nil, err
			}
			if err := rn.PipeE(yaml.SetAnnotation(kioutil.IndexAnnotation, "new")); err != nil {
				return nil, err
			}
		}
		return nodes, nil
	}
	changeLegacyAnnos := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, rn := range nodes {
			if err := rn.PipeE(yaml.SetAnnotation(kioutil.LegacyPathAnnotation, "new")); err != nil {
				return nil, err
			}
			if err := rn.PipeE(yaml.SetAnnotation(kioutil.LegacyIndexAnnotation, "new")); err != nil {
				return nil, err
			}
		}
		return nodes, nil
	}
	changeLegacyId := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, rn := range nodes {
			if err := rn.PipeE(yaml.SetAnnotation(kioutil.LegacyIdAnnotation, "new")); err != nil {
				return nil, err
			}
		}
		return nodes, nil
	}
	changeInternalId := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, rn := range nodes {
			if err := rn.PipeE(yaml.SetAnnotation(kioutil.IdAnnotation, "new")); err != nil {
				return nil, err
			}
		}
		return nodes, nil
	}
	changeBothPathAnnos := func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, rn := range nodes {
			if err := rn.PipeE(yaml.SetAnnotation(kioutil.LegacyPathAnnotation, "legacy")); err != nil {
				return nil, err
			}
			if err := rn.PipeE(yaml.SetAnnotation(kioutil.PathAnnotation, "new")); err != nil {
				return nil, err
			}
		}
		return nodes, nil
	}

	noops := []Filter{
		FilterFunc(noopFilter1),
		FilterFunc(noopFilter2),
	}
	internal := []Filter{FilterFunc(changeInternalAnnos)}
	legacy := []Filter{FilterFunc(changeLegacyAnnos)}
	legacyId := []Filter{FilterFunc(changeLegacyId)}
	internalId := []Filter{FilterFunc(changeInternalId)}
	changeBoth := []Filter{FilterFunc(changeBothPathAnnos), FilterFunc(noopFilter1)}

	testCases := map[string]struct {
		input       string
		filters     []Filter
		expected    string
		expectedErr string
	}{
		// the orchestrator should copy the legacy annotations to the new
		// annotations
		"legacy annotations only": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '0'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    config.kubernetes.io/path: "configmap.yaml"
    config.kubernetes.io/index: '1'
data:
  grpcPort: 8081
`,
			filters: noops,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'configmap.yaml'
    internal.config.kubernetes.io/index: '0'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    config.kubernetes.io/path: "configmap.yaml"
    config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'configmap.yaml'
    internal.config.kubernetes.io/index: '1'
data:
  grpcPort: 8081
`,
		},
		// the orchestrator should copy the new annotations to the
		// legacy annotations
		"new annotations only": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    internal.config.kubernetes.io/path: 'configmap.yaml'
    internal.config.kubernetes.io/index: '0'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    internal.config.kubernetes.io/path: "configmap.yaml"
    internal.config.kubernetes.io/index: '1'
data:
  grpcPort: 8081
`,
			filters: noops,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    internal.config.kubernetes.io/path: 'configmap.yaml'
    internal.config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '0'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    internal.config.kubernetes.io/path: "configmap.yaml"
    internal.config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '1'
data:
  grpcPort: 8081
`,
		},
		// the orchestrator should detect that the legacy annotations
		// have been changed by the function, and should update the
		// new internal annotations to reflect the same change
		"change only legacy annotations": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '0'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    config.kubernetes.io/path: "configmap.yaml"
    config.kubernetes.io/index: '1'
data:
  grpcPort: 8081
`,
			filters: legacy,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'new'
    config.kubernetes.io/index: 'new'
    internal.config.kubernetes.io/path: 'new'
    internal.config.kubernetes.io/index: 'new'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    config.kubernetes.io/path: "new"
    config.kubernetes.io/index: 'new'
    internal.config.kubernetes.io/path: 'new'
    internal.config.kubernetes.io/index: 'new'
data:
  grpcPort: 8081
`,
		},
		// the orchestrator should detect that the new internal annotations
		// have been changed by the function, and should update the
		// legacy annotations to reflect the same change
		"change only internal annotations": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '0'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    config.kubernetes.io/path: "configmap.yaml"
    config.kubernetes.io/index: '1'
data:
  grpcPort: 8081
`,
			filters: internal,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'new'
    config.kubernetes.io/index: 'new'
    internal.config.kubernetes.io/path: 'new'
    internal.config.kubernetes.io/index: 'new'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    config.kubernetes.io/path: "new"
    config.kubernetes.io/index: 'new'
    internal.config.kubernetes.io/path: 'new'
    internal.config.kubernetes.io/index: 'new'
data:
  grpcPort: 8081
`,
		},
		// the orchestrator should detect that the new internal id annotation
		// has been changed, and copy it over to the legacy one, and also
		// copy the path and index legacy annotations to the new internal
		// ones
		"change only internal id when original legacy set": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '0'
    config.k8s.io/id: '1'
data:
  grpcPort: 8080
`,
			filters: internalId,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '0'
    config.k8s.io/id: 'new'
    internal.config.kubernetes.io/path: 'configmap.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/id: 'new'
data:
  grpcPort: 8080
`,
		},
		// the orchestrator should detect that the legacy id annotation
		// has been changed, and copy it over to the internal one, and also
		// copy the path and index internal annotations to the legacy
		// ones
		"change only legacy id when internal legacy set": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    internal.config.kubernetes.io/path: 'configmap.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/id: '1'
data:
  grpcPort: 8080
`,
			filters: legacyId,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    internal.config.kubernetes.io/path: 'configmap.yaml'
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/id: 'new'
    config.kubernetes.io/path: 'configmap.yaml'
    config.kubernetes.io/index: '0'
    config.k8s.io/id: 'new'
data:
  grpcPort: 8080
`,
		},
		// the function changes both the legacy and new path annotation,
		// so we should get an error
		"change both": {
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-from
  annotations:
    config.kubernetes.io/path: 'configmap.yaml'
    internal.kubernetes.io/path: 'configmap.yaml'
data:
  grpcPort: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ports-to
  annotations:
    config.kubernetes.io/path: "configmap.yaml"
    config.kubernetes.io/index: '1'
data:
  grpcPort: 8081
`,
			filters:     changeBoth,
			expectedErr: "resource input to function has mismatched legacy and internal path annotations",
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			var out bytes.Buffer
			input := ByteReadWriter{
				Reader:                bytes.NewBufferString(tc.input),
				Writer:                &out,
				OmitReaderAnnotations: true,
				KeepReaderAnnotations: true,
			}
			p := Pipeline{
				Inputs:  []Reader{&input},
				Filters: tc.filters,
				Outputs: []Writer{&input},
			}

			err := p.Execute()
			if tc.expectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, out.String())
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err.Error())
			}

		})
	}
}
