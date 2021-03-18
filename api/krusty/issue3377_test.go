// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestIssue3377(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	opts := th.MakeDefaultOptions()
	th.WriteK(".", `
resources:
- service-a.yaml
- service-b.yaml

patchesJson6902:
- path: service-a-patch.yaml
  target:
    version: v1
    group: networking.k8s.io
    kind: Ingress
    name: service-a
- path: service-b-patch.yaml
  target:
    version: v1
    group: networking.k8s.io
    kind: Ingress
    name: service-b
vars:
- name: SERVICE_A_DNS_NAME
  objref:
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    name: service-a
  fieldref:
    fieldpath: spec.rules[0].host
- name: SERVICE_B_DNS_NAME
  objref:
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    name: service-b
  fieldref:
    fieldpath: spec.rules[0].host
`)

	th.WriteF("service-a.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service-a
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: service-a
  template:
    metadata:
      labels:
        app.kubernetes.io/component: service-a
    spec:
      containers:
      - image: repository/service-a:v1
        imagePullPolicy: Always
        name: service-b
        ports:
        - containerPort: 8080
        env:
        - name: REDIRECT_URI
          value: "http://$(SERVICE_B_DNS_NAME):80/oauth_redir"
---
apiVersion: v1
kind: Service
metadata:
  name: service-a
  labels:
    app.kubernetes.io/component: service-a
spec:
  type: LoadBalancer
  ports:
  - port: 8080
  selector:
    app.kubernetes.io/component: service-a
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: service-a
spec:
  rules:
  - host: service-a.k8s.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: service-a
            port:
              number: 8080
`)

	th.WriteF("service-a-patch.yaml", `
- op: replace
  path: /spec/rules/0/host
  value: new-service-a.k8s.com
`)

	th.WriteF("service-b.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service-b
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: service-b
  template:
    metadata:
      labels:
        app.kubernetes.io/component: service-b
    spec:
      containers:
      - image: repository/service-b:v1
        imagePullPolicy: Always
        name: service-b
        ports:
        - containerPort: 8080
        env:
        - name: REDIRECT_URI
          value: "http://$(SERVICE_A_DNS_NAME):80/oauth_redir"
---
apiVersion: v1
kind: Service
metadata:
  name: service-b
  labels:
    app.kubernetes.io/component: service-b
spec:
  type: LoadBalancer
  ports:
  - port: 8080
  selector:
    app.kubernetes.io/component: service-b
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: service-b
spec:
  rules:
  - host: service-b.k8s.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: service-b
            port:
              number: 8080
`)
	th.WriteF("service-b-patch.yaml", `
- op: replace
  path: /spec/rules/0/host
  value: new-service-b.k8s.com
`)
	m := th.Run(".", opts)
	th.AssertActualEqualsExpected(
		m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service-a
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: service-a
  template:
    metadata:
      labels:
        app.kubernetes.io/component: service-a
    spec:
      containers:
      - env:
        - name: REDIRECT_URI
          value: http://new-service-b.k8s.com:80/oauth_redir
        image: repository/service-a:v1
        imagePullPolicy: Always
        name: service-b
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: service-a
  name: service-a
spec:
  ports:
  - port: 8080
  selector:
    app.kubernetes.io/component: service-a
  type: LoadBalancer
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: service-a
spec:
  rules:
  - host: new-service-a.k8s.com
    http:
      paths:
      - backend:
          service:
            name: service-a
            port:
              number: 8080
        path: /
        pathType: Prefix
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service-b
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: service-b
  template:
    metadata:
      labels:
        app.kubernetes.io/component: service-b
    spec:
      containers:
      - env:
        - name: REDIRECT_URI
          value: http://new-service-a.k8s.com:80/oauth_redir
        image: repository/service-b:v1
        imagePullPolicy: Always
        name: service-b
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: service-b
  name: service-b
spec:
  ports:
  - port: 8080
  selector:
    app.kubernetes.io/component: service-b
  type: LoadBalancer
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: service-b
spec:
  rules:
  - host: new-service-b.k8s.com
    http:
      paths:
      - backend:
          service:
            name: service-b
            port:
              number: 8080
        path: /
        pathType: Prefix
`)
}
