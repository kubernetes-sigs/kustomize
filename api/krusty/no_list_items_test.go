// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// test for https://github.com/kubernetes-sigs/kustomize/issues/4240
func TestSuffix5042(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- resource.yaml
`)

	th.WriteF("resource.yaml", `
apiVersion: example.com/v1alpha1
kind: MyResource
metadata:
  name: service
---
apiVersion: example.com/v1alpha1
kind: MyResourceTwo
metadata:
  name: test
rules: []
`)
	m := th.Run(".", th.MakeDefaultOptions())

	th.AssertActualEqualsExpected(m, `
apiVersion: example.com/v1alpha1
kind: MyResource
metadata:
  name: service
---
apiVersion: example.com/v1alpha1
kind: MyResourceTwo
metadata:
  name: test
rules: []
`)
}

func TestListSuffix5042(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- resource.yaml
`)

	th.WriteF("resource.yaml", `
apiVersion: example.com/v1alpha1
kind: MyResource
metadata:
  name: service
---
apiVersion: example.com/v1alpha1
kind: MyResourceList
metadata:
  name: test
rules: []
`)
	m := th.Run(".", th.MakeDefaultOptions())

	th.AssertActualEqualsExpected(m, `
apiVersion: example.com/v1alpha1
kind: MyResource
metadata:
  name: service
---
apiVersion: example.com/v1alpha1
kind: MyResourceList
metadata:
  name: test
rules: []
`)
}

func TestListSuffix5485(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- resource.yaml
  `)

	th.WriteF("resource.yaml", `
apiVersion: infra.protonbase.io/v1alpha1
kind: AccessWhiteList
metadata:
  name: wlmls5769f
  namespace: dc7i4hyxzw
spec:
  rules:
  - sourceIps: 0.0.0.0/16
`)

	m := th.Run(".", th.MakeDefaultOptions())

	th.AssertActualEqualsExpected(m, `
apiVersion: infra.protonbase.io/v1alpha1
kind: AccessWhiteList
metadata:
  name: wlmls5769f
  namespace: dc7i4hyxzw
spec:
  rules:
  - sourceIps: 0.0.0.0/16
`)
}

func TestListToIndividualResources(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- list.yaml
`)

	th.WriteF("list.yaml", `
apiVersion: v1
kind: PodList
items:
  - apiVersion: v1
    kind: Pod
    metadata:
      name: my-pod-1
      namespace: default
      labels:
        app: my-app
    spec:
      containers:
        - name: my-container
          image: nginx:1.19.10
          ports:
            - containerPort: 80
  - apiVersion: v1
    kind: Pod
    metadata:
      name: my-pod-2
      namespace: default
      labels:
        app: my-app
    spec:
      containers:
        - name: my-container
          image: nginx:1.19.10
          ports:
            - containerPort: 80
  - apiVersion: v1
    kind: Pod
    metadata:
      name: my-pod-3
      namespace: default
      labels:
        app: my-app
    spec:
      containers:
        - name: my-container
          image: nginx:1.19.10
          ports:
            - containerPort: 80
`)

	m := th.Run(".", th.MakeDefaultOptions())

	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: my-app
  name: my-pod-1
  namespace: default
spec:
  containers:
  - image: nginx:1.19.10
    name: my-container
    ports:
    - containerPort: 80
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: my-app
  name: my-pod-2
  namespace: default
spec:
  containers:
  - image: nginx:1.19.10
    name: my-container
    ports:
    - containerPort: 80
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: my-app
  name: my-pod-3
  namespace: default
spec:
  containers:
  - image: nginx:1.19.10
    name: my-container
    ports:
    - containerPort: 80
`)
}

// Empty list should result in no resources
func TestEmptyList(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- emptyList.yaml
`)
	th.WriteF("emptyList.yaml", `
apiVersion: v1
kind: PodList
items: []
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, "")
}
