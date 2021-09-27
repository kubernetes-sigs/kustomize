// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements adding an Application CR to a group of resources and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

var input = `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: the-map
    annotations:
      config.kubernetes.io/index: '0'
      config.kubernetes.io/path: 'example.yaml'
  data:
    altGreeting: "Good Morning!"
    enableRisky: "false"
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: the-deployment
    annotations:
      config.kubernetes.io/index: '1'
      config.kubernetes.io/path: 'example.yaml'
  spec:
    replicas: 3
    template:
      metadata:
        labels:
          deployment: hello
      spec:
        containers:
        - name: the-container
          image: monopole/hello:1
          command: ["/hello", "--port=8080", "--enableRiskyFeature=$(ENABLE_RISKY)"]
          ports:
          - containerPort: 8080
          env:
          - name: ALT_GREETING
            valueFrom:
              configMapKeyRef:
                name: the-map
                key: altGreeting
          - name: ENABLE_RISKY
            valueFrom:
              configMapKeyRef:
                name: the-map
                key: enableRisky
- kind: Service
  apiVersion: v1
  metadata:
    name: the-service
    annotations:
      config.kubernetes.io/index: '2'
      config.kubernetes.io/path: 'example.yaml'
  spec:
    selector:
      deployment: hello
    type: LoadBalancer
    ports:
    - protocol: TCP
      port: 8666
      targetPort: 8080
functionConfig:
  # Copyright 2019 The Kubernetes Authors.
  # SPDX-License-Identifier: Apache-2.0
  apiVersion: examples.config.kubernetes.io/v1beta1
  kind: CreateApp
  spec:
    managedBy: jingfang
    name: example-app
    namespace: example-namespace
    descriptor:
      links:
      - description: About
        url: https://worldpress.org/
      - description: web server dashboard
        url: https://metrics/internal/worldpress-01/web-app
`

var output = `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: the-map
    annotations:
      config.kubernetes.io/index: '0'
      config.kubernetes.io/path: 'example.yaml'
    labels:
      app.kubernetes.io/name: 'example-app'
  data:
    altGreeting: "Good Morning!"
    enableRisky: "false"
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: the-deployment
    annotations:
      config.kubernetes.io/index: '1'
      config.kubernetes.io/path: 'example.yaml'
    labels:
      app.kubernetes.io/name: 'example-app'
  spec:
    replicas: 3
    template:
      metadata:
        labels:
          deployment: hello
      spec:
        containers:
        - name: the-container
          image: monopole/hello:1
          command: ["/hello", "--port=8080", "--enableRiskyFeature=$(ENABLE_RISKY)"]
          ports:
          - containerPort: 8080
          env:
          - name: ALT_GREETING
            valueFrom:
              configMapKeyRef:
                name: the-map
                key: altGreeting
          - name: ENABLE_RISKY
            valueFrom:
              configMapKeyRef:
                name: the-map
                key: enableRisky
- kind: Service
  apiVersion: v1
  metadata:
    name: the-service
    annotations:
      config.kubernetes.io/index: '2'
      config.kubernetes.io/path: 'example.yaml'
    labels:
      app.kubernetes.io/name: 'example-app'
  spec:
    selector:
      deployment: hello
    type: LoadBalancer
    ports:
    - protocol: TCP
      port: 8666
      targetPort: 8080
- apiVersion: app.k8s.io/v1beta1
  kind: Application
  metadata:
    annotations:
      app.kubernetes.io/managed-by: jingfang
    creationTimestamp: null
    labels:
      app.kubernetes.io/name: example-app
    name: example-app
    namespace: example-namespace
  spec:
    componentKinds:
    - group: ""
      kind: ConfigMap
    - group: apps
      kind: Deployment
    - group: ""
      kind: Service
    descriptor:
      links:
      - description: About
        url: https://worldpress.org/
      - description: web server dashboard
        url: https://metrics/internal/worldpress-01/web-app
    selector:
      matchLabels:
        app.kubernetes.io/name: example-app
  status: {}
functionConfig:
  # Copyright 2019 The Kubernetes Authors.
  # SPDX-License-Identifier: Apache-2.0
  apiVersion: examples.config.kubernetes.io/v1beta1
  kind: CreateApp
  spec:
    managedBy: jingfang
    name: example-app
    namespace: example-namespace
    descriptor:
      links:
      - description: About
        url: https://worldpress.org/
      - description: web server dashboard
        url: https://metrics/internal/worldpress-01/web-app
`

func Test(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			outC <- ""
		}
		outC <- buf.String()
	}()

	tmpfile, err := ioutil.TempFile("", "test-input")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // noerrcheck
	if _, err := tmpfile.Write([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	os.Stdin = tmpfile

	err = appCR()
	if err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	out := <-outC
	if out != output {
		t.Fatalf("expected %s\nbut got %s\n", output, out)
	}
}
