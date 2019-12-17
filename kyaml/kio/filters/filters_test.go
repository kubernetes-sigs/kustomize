// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
	. "sigs.k8s.io/kustomize/kyaml/kio/filters"
)

var r = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1
  namespace: bar
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo2
---
apiVersion: v1
kind: Service
metadata:
  name: foo2
  namespace: bar
---
apiVersion: v1
kind: Service
metadata:
  name: foo1
`

func TestFileSetter_Filter(t *testing.T) {
	in := bytes.NewBufferString(r)
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs:  []Reader{&ByteReader{Reader: in}},
		Filters: []Filter{&FileSetter{}},
		Outputs: []Writer{ByteWriter{Sort: true, Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1
  namespace: bar
  annotations:
    config.kubernetes.io/path: 'foo1_deployment.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: foo1
  annotations:
    config.kubernetes.io/path: 'foo1_service.yaml'
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo2
  annotations:
    config.kubernetes.io/path: 'foo2_deployment.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: foo2
  namespace: bar
  annotations:
    config.kubernetes.io/path: 'foo2_service.yaml'
`, out.String())
}

func TestFileSetter_Filter_pattern(t *testing.T) {
	in := bytes.NewBufferString(r)
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs: []Reader{&ByteReader{Reader: in}},
		Filters: []Filter{&FileSetter{
			FilenamePattern: "%n_%s_%k.yaml",
		}},
		Outputs: []Writer{ByteWriter{Sort: true, Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `apiVersion: v1
kind: Service
metadata:
  name: foo1
  annotations:
    config.kubernetes.io/path: 'foo1__service.yaml'
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1
  namespace: bar
  annotations:
    config.kubernetes.io/path: 'foo1_bar_deployment.yaml'
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo2
  annotations:
    config.kubernetes.io/path: 'foo2__deployment.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: foo2
  namespace: bar
  annotations:
    config.kubernetes.io/path: 'foo2_bar_service.yaml'
`, out.String())
}

func TestFileSetter_Filter_empty(t *testing.T) {
	in := bytes.NewBufferString(r)
	out := &bytes.Buffer{}
	err := Pipeline{
		Inputs: []Reader{&ByteReader{Reader: in}},
		Filters: []Filter{&FileSetter{
			FilenamePattern: "resource.yaml",
		}},
		Outputs: []Writer{ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1
  namespace: bar
  annotations:
    config.kubernetes.io/path: 'resource.yaml'
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo2
  annotations:
    config.kubernetes.io/path: 'resource.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: foo2
  namespace: bar
  annotations:
    config.kubernetes.io/path: 'resource.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: foo1
  annotations:
    config.kubernetes.io/path: 'resource.yaml'
`, out.String())
}
