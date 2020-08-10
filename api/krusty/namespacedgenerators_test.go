// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestNamespacedGenerator(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: the-non-default-namespace-map
  namespace: non-default
  literals:
  - altGreeting=Good Morning from non-default namespace!
  - enableRisky="false"
- name: the-map
  literals:
  - altGreeting=Good Morning from default namespace!
  - enableRisky="false"

secretGenerator:
- name: the-non-default-namespace-secret
  namespace: non-default
  literals:
  - password.txt=verySecret
- name: the-secret
  literals:
  - password.txt=anotherSecret
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  altGreeting: Good Morning from non-default namespace!
  enableRisky: "false"
kind: ConfigMap
metadata:
  name: the-non-default-namespace-map-64b2md8tth
  namespace: non-default
---
apiVersion: v1
data:
  altGreeting: Good Morning from default namespace!
  enableRisky: "false"
kind: ConfigMap
metadata:
  name: the-map-tg7t5hk8bk
---
apiVersion: v1
data:
  password.txt: dmVyeVNlY3JldA==
kind: Secret
metadata:
  name: the-non-default-namespace-secret-8tc9gdd76t
  namespace: non-default
type: Opaque
---
apiVersion: v1
data:
  password.txt: YW5vdGhlclNlY3JldA==
kind: Secret
metadata:
  name: the-secret-6557m7fcg8
type: Opaque
`)
}

func TestNamespacedGeneratorWithOverlays(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
namespace: base

configMapGenerator:
- name: testCase
  literals:
    - base=true
`)
	th.WriteK("/app/overlay", `
resources:
  - ../base

namespace: overlay

configMapGenerator:
  - name: testCase
    behavior: merge
    literals:
      - overlay=true
`)
	m := th.Run("/app/overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  base: "true"
  overlay: "true"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: testCase-bcbmmg48hd
  namespace: overlay
`)
}
