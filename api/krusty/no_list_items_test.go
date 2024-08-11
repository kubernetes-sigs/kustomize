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
  annotations:
    kustomize.config.k8s.io/convertToInlineList: "false"
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
  annotations:
    kustomize.config.k8s.io/convertToInlineList: "false"
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
  annotations:
    kustomize.config.k8s.io/convertToInlineList: "false"
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
  annotations:
    kustomize.config.k8s.io/convertToInlineList: "false"
  name: wlmls5769f
  namespace: dc7i4hyxzw
spec:
  rules:
  - sourceIps: 0.0.0.0/16
`)
  
}