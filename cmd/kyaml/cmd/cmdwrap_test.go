// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	input = `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
functionConfig:
  metadata:
    name: test
  spec:
    replicas: 11
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: test
    labels:
      app: nginx
      name: test
  spec:
    replicas: 5
    selector:
      matchLabels:
        app: nginx
        name: test
    template:
      metadata:
        labels:
          app: nginx
          name: test
      spec:
        containers:
        - name: test
          image: nginx:v1.7
          ports:
          - containerPort: 8080
            name: http
          resources:
            limits:
              cpu: 500m
- apiVersion: v1
  kind: Service
  metadata:
    name: test
    labels:
      app: nginx
      name: test
  spec:
    ports:
    # This i the port.
    - port: 8080
      targetPort: 8080
      name: http
    selector:
      app: nginx
      name: test
`

	output = `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: test
    labels:
      name: test
      app: nginx
    annotations:
      config.kubernetes.io/index: 0
      config.kubernetes.io/path: config/test_deployment.yaml
  spec:
    replicas: 11
    selector:
      matchLabels:
        name: test
        app: nginx
    template:
      metadata:
        labels:
          name: test
          app: nginx
      spec:
        containers:
        - name: test
          image: nginx:v1.7
          ports:
          - name: http
            containerPort: 8080
          resources:
            limits:
              cpu: 500m
- apiVersion: v1
  kind: Service
  metadata:
    name: test
    labels:
      name: test
      app: nginx
    annotations:
      config.kubernetes.io/index: 0
      config.kubernetes.io/path: config/test_service.yaml
  spec:
    selector:
      name: test
      app: nginx
    ports:
    - name: http
      # This i the port.
      port: 8080
      targetPort: 8080
`

	outputNoMerge = `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: test
    labels:
      name: test
      app: nginx
    annotations:
      config.kubernetes.io/index: 0
      config.kubernetes.io/path: config/test_deployment.yaml
  spec:
    replicas: 11
    selector:
      matchLabels:
        name: test
        app: nginx
    template:
      metadata:
        labels:
          name: test
          app: nginx
      spec:
        containers:
        - name: test
          image: nginx:v1.7
          ports:
          - name: http
            containerPort: 8080
- apiVersion: v1
  kind: Service
  metadata:
    name: test
    labels:
      name: test
      app: nginx
    annotations:
      config.kubernetes.io/index: 0
      config.kubernetes.io/path: config/test_service.yaml
  spec:
    selector:
      name: test
      app: nginx
    ports:
    - name: http
      # This i the port.
      port: 8080
      targetPort: 8080
`

	outputOverride = `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: test
    labels:
      name: test
      app: nginx
    annotations:
      config.kubernetes.io/index: 0
      config.kubernetes.io/path: config/test_deployment.yaml
  spec:
    replicas: 11
    selector:
      matchLabels:
        name: test
        app: nginx
    template:
      metadata:
        labels:
          name: test
          app: nginx
      spec:
        containers:
        - name: test
          image: nginx:v1.9
          ports:
          - name: http
            containerPort: 8080
          resources:
            limits:
              cpu: 500m
- apiVersion: v1
  kind: Service
  metadata:
    name: test
    labels:
      name: test
      app: nginx
    annotations:
      config.kubernetes.io/index: 0
      config.kubernetes.io/path: config/test_service.yaml
  spec:
    selector:
      name: test
      app: nginx
    ports:
    - name: http
      # This i the port.
      port: 8080
      targetPort: 8080
`
)

func TestCmd_wrap(t *testing.T) {
	_, dir, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	dir = filepath.Dir(dir)

	c := GetWrapRunner()
	c.Command.SetIn(bytes.NewBufferString(input))
	out := &bytes.Buffer{}
	c.Command.SetOut(out)
	args := []string{"--", filepath.Join(dir, "test", "test.sh")}
	c.Command.SetArgs(args)
	c.XArgs.Args = args

	if !assert.NoError(t, c.Command.Execute()) {
		t.FailNow()
	}

	assert.Equal(t, output, out.String())
}

func TestCmd_wrapNoMerge(t *testing.T) {
	_, dir, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	dir = filepath.Dir(dir)

	c := GetWrapRunner()
	c.getEnv = func(key string) string {
		if key == KustMergeEnv {
			return "false"
		}
		return ""
	}
	c.Command.SetIn(bytes.NewBufferString(input))
	out := &bytes.Buffer{}
	c.Command.SetOut(out)
	args := []string{"--", filepath.Join(dir, "test", "test.sh")}
	c.Command.SetArgs(args)
	c.XArgs.Args = args
	if !assert.NoError(t, c.Command.Execute()) {
		t.FailNow()
	}

	assert.Equal(t, outputNoMerge, out.String())
}

func TestCmd_wrapOverride(t *testing.T) {
	_, dir, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	dir = filepath.Dir(dir)

	c := GetWrapRunner()
	c.getEnv = func(key string) string {
		if key == KustOverrideDirEnv {
			return filepath.Join(dir, "test")
		}
		return ""
	}
	c.Command.SetIn(bytes.NewBufferString(input))
	out := &bytes.Buffer{}
	c.Command.SetOut(out)
	args := []string{"--", filepath.Join(dir, "test", "test.sh")}
	c.Command.SetArgs(args)
	c.XArgs.Args = args
	if !assert.NoError(t, c.Command.Execute()) {
		t.FailNow()
	}

	assert.Equal(t, outputOverride, out.String())
}
