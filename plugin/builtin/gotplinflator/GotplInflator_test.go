// +build notravis

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// FIXME, Disabled on travis, because don't want to install go-getter on travis.

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// FIXME, This test requires having the go-getter binary on the PATH.
func TestGotplInflator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
    PrepBuiltin("GotplInflator")
	defer th.Reset()

	m := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: GotplInflator
metadata:
  name: example

dependencies:
- name: nginx
  repo: https://github.com/epcim/k8s-manifests?ref=main
  path: example/manifests
  #pull: Always
  #templateGlob:   "*.t*pl"
  #templateOpts    # PLACEHOLDER

values:
  nginx_cpu_request: "512m"
  nginx:
    cpu:
      limit:  "1000m"
    memory:
      limit:  "1024M"

`)

	th.AssertActualEqualsExpected(m, `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
    component: web
    role: front
  name: nginx
  namespace: web
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
      component: web
      name: nginx
      role: front
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        checksum/config: unknown
      labels:
        app: nginx
        component: web
        name: nginx
        role: front
    spec:
      containers:
      - image: nginx:stable
        imagePullPolicy: IfNotPresent
        name: nginx
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        resources:
          limits:
            cpu: 1000m
            memory: 1024M
          requests:
            cpu: 512m
            memory: 32Mi
      imagePullSecrets:
      - name: null
      serviceAccountName: nginx
      volumes:
      - configMap:
          name: nginx-config
        name: nginx-config
      - emptyDir:
          medium: Memory
          sizeLimit: 5M
        name: certs-volume
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
    component: web
    role: front
  name: nginx
  namespace: web
spec:
  ports:
  - port: 8080
    protocol: TCP
  selector:
    app: nginx
    component: nginx
---
apiVersion: v1
data:
  nginx-config.yaml: '# DUMMY'
kind: ConfigMap
metadata:
  labels:
    app: nginx
  name: nginx-config
  namespace: web
`)
}
