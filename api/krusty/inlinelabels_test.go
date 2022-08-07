// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const resources string = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      containers:
      - name: my-deployment
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
---
apiVersion: example.dev/v1
kind: MyCRD
metadata:
  name: crd
`

func TestKustomizationLabels(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeResourcesForPatchTest(th)
	th.WriteK("/app", `
resources:
- deployment.yaml

labels:
- pairs:
    foo: bar
- pairs:
    a: b
  includeSelectors: true
- pairs:
    c: d
  fields:
  - path: spec/selector
    group: example.dev
    version: v1
    kind: MyCRD
    create: true
`)
	th.WriteF("/app/deployment.yaml", resources)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    a: b
    c: d
    foo: bar
  name: my-deployment
spec:
  selector:
    matchLabels:
      a: b
  template:
    metadata:
      labels:
        a: b
    spec:
      containers:
      - livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
        name: my-deployment
---
apiVersion: example.dev/v1
kind: MyCRD
metadata:
  labels:
    a: b
    c: d
    foo: bar
  name: crd
spec:
  selector:
    c: d
`)
}

func TestKustomizationLabelsInTemplate(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: a
    app.kubernetes.io/instance: b
    app.kubernetes.io/name: c
    app.kubernetes.io/part-of: d
  name: deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: a
      app.kubernetes.io/instance: b
      app.kubernetes.io/name: c
      app.kubernetes.io/part-of: d
  template:
    metadata:
      labels:
        app.kubernetes.io/component: a
        app.kubernetes.io/instance: b
        app.kubernetes.io/name: c
        app.kubernetes.io/part-of: d
`)
	th.WriteK("/app", `
resources:
- deployment.yaml

labels:
- pairs:
    foo: bar
  includeSelectors: false
  includeTemplates: true
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: a
    app.kubernetes.io/instance: b
    app.kubernetes.io/name: c
    app.kubernetes.io/part-of: d
    foo: bar
  name: deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: a
      app.kubernetes.io/instance: b
      app.kubernetes.io/name: c
      app.kubernetes.io/part-of: d
  template:
    metadata:
      labels:
        app.kubernetes.io/component: a
        app.kubernetes.io/instance: b
        app.kubernetes.io/name: c
        app.kubernetes.io/part-of: d
        foo: bar
`)
}

func TestKustomizationLabelsInTemplateWhenLabelsIsNil(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: test-server
        image: test-server
`)
	th.WriteK("/app", `
resources:
- deployment.yaml

commonLabels:
  app: test-server

labels:
- pairs:
    app.kubernetes.io/component: a
    app.kubernetes.io/instance: b
    app.kubernetes.io/name: c
    app.kubernetes.io/part-of: d
  includeSelectors: false
  includeTemplates: true
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-server
    app.kubernetes.io/component: a
    app.kubernetes.io/instance: b
    app.kubernetes.io/name: c
    app.kubernetes.io/part-of: d
  name: deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-server
  template:
    metadata:
      labels:
        app: test-server
        app.kubernetes.io/component: a
        app.kubernetes.io/instance: b
        app.kubernetes.io/name: c
        app.kubernetes.io/part-of: d
    spec:
      containers:
      - image: test-server
        name: test-server
`)
}
