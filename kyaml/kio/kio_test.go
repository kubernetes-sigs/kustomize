// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Some test configs
var deployment = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo2
  annotations:
    foo: bar
`

var service = `apiVersion: v1
kind: Service
metadata:
  name: the-service
spec:
  selector:
    deployment: hello
  type: LoadBalancer
  ports:
  - protocol: TCP
    port: 8666
    targetPort: 8080
`

// Config with "no-process" annotation, which should be skipped
// in pipeline execution
var noProcessAnnotation = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  namespace: test-namespace
  annotations:
    config.kubernetes.io/no-process: true
`

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

func TestSlice_Write(t *testing.T) {

}

// "No-process" annotation skips config in pipeline execute.
func TestNoProcessForPipelineNodes(t *testing.T) {
	// Test function to add a "foo: 'bar'" annotation to a config.
	setAnnotationFn := FilterFunc(func(operand []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range operand {
			resource := operand[i]
			_, err := resource.Pipe(yaml.SetAnnotation("foo", "bar"))
			if err != nil {
				return nil, err
			}
		}
		return operand, nil
	})
	// Service config should have foo:bar annotation added
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs: []Reader{
			&ByteReader{Reader: bytes.NewBufferString(service)},
		},
		// Service should have annotation set with filter
		Filters: []Filter{setAnnotationFn},
		Outputs: []Writer{ByteWriter{Sort: true, Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		return
	}
	if !strings.Contains(out.String(), "foo: 'bar'") {
		t.Errorf("expected service with annotation got (%s)", out.String())
	}
	// Deployment with no-process annotation should not have annotation set.
	out = &bytes.Buffer{}
	err = Pipeline{
		Inputs: []Reader{
			&ByteReader{Reader: bytes.NewBufferString(noProcessAnnotation)},
		},
		Filters: []Filter{setAnnotationFn},
		Outputs: []Writer{ByteWriter{Sort: true, Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, noProcessAnnotation, out.String())
}

func TestIsNoProcessNode(t *testing.T) {
	// Create three test nodes: one with the no-process annotation,
	// one with annotations (but without the no-process annotation),
	// and one with no annotations.
	noProcessNode, err := yaml.Parse(noProcessAnnotation)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	annotationsNode, err := yaml.Parse(deployment)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	simpleNode, err := yaml.Parse(`e: f`)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	testCases := map[string]struct {
		node          *yaml.RNode
		isProcessNode bool
	}{
		"nil RNode is not process node": {
			node:          nil,
			isProcessNode: false,
		},
		"Empty RNode is not process node": {
			node:          &yaml.RNode{},
			isProcessNode: false,
		},
		"Simple RNode without annotation is not process node": {
			node:          simpleNode,
			isProcessNode: false,
		},
		"RNode with annotations, but not no-process is not process node": {
			node:          annotationsNode,
			isProcessNode: false,
		},
		"RNode with no-process annotation is process node": {
			node:          noProcessNode,
			isProcessNode: true,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			actual := IsNoProcessNode(tc.node)
			if tc.isProcessNode != actual {
				t.Errorf("isNoProcessNode expected (%t), got (%t)", tc.isProcessNode, actual)
			}
		})
	}
}
