// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSecretGenerator(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
secretGenerator:
- name: bob
  literals:
  - FRUIT=apple
  - VEGETABLE=carrot
  files:
  - foo.env
  - passphrase=phrase.dat
  envs:
  - foo.env
`)
	th.WriteF("foo.env", `
MOUNTAIN=everest
OCEAN=pacific
`)
	th.WriteF("phrase.dat", "dat phrase")
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  FRUIT: YXBwbGU=
  MOUNTAIN: ZXZlcmVzdA==
  OCEAN: cGFjaWZpYw==
  VEGETABLE: Y2Fycm90
  foo.env: Ck1PVU5UQUlOPWV2ZXJlc3QKT0NFQU49cGFjaWZpYwo=
  passphrase: ZGF0IHBocmFzZQ==
kind: Secret
metadata:
  name: bob-bh645k7tmg
type: Opaque
`)
}

func TestGeneratorOptionsWithBases(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
generatorOptions:
  disableNameSuffixHash: true
  labels:
    foo: bar
configMapGenerator:
- name: shouldNotHaveHash
  literals:
  - foo=bar
`)
	th.WriteK("overlay", `
resources:
- ../base
generatorOptions:
  disableNameSuffixHash: false
  labels:
    fruit: apple
configMapGenerator:
- name: shouldHaveHash
  literals:
  - fruit=apple
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    foo: bar
  name: shouldNotHaveHash
---
apiVersion: v1
data:
  fruit: apple
kind: ConfigMap
metadata:
  labels:
    fruit: apple
  name: shouldHaveHash-c9867f8446
`)
}

func TestGeneratorOptionsOverlayDisableNameSuffixHash(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
generatorOptions:
  disableNameSuffixHash: false
  labels:
    foo: bar
configMapGenerator:
- name: baseShouldHaveHashButOverlayShouldNot
  literals:
  - foo=bar
`)
	th.WriteK("overlay", `
resources:
- ../base
generatorOptions:
  disableNameSuffixHash: true
  labels:
    fruit: apple
configMapGenerator:
- name: baseShouldHaveHashButOverlayShouldNot
  behavior: merge
  literals:
  - fruit=apple
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  foo: bar
  fruit: apple
kind: ConfigMap
metadata:
  labels:
    foo: bar
    fruit: apple
  name: baseShouldHaveHashButOverlayShouldNot
`)
}
