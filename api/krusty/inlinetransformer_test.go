// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestInlineTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("resource.yaml", `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
data: {}
`)
	th.WriteK(".", `
resources:
- resource.yaml
transformers:
- |-
  apiVersion: builtin
  kind: NamespaceTransformer
  metadata:
    name: not-important-to-example
    namespace: test
  fieldSpecs:
  - path: metadata/namespace
    create: true
`)

	expected := `
apiVersion: apps/v1
data: {}
kind: ConfigMap
metadata:
  name: whatever
  namespace: test
`

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}

func TestInlineGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK(".", `
generators:
- |-
  apiVersion: builtin
  kind: ConfigMapGenerator
  metadata:
    name: mymap
  literals:
  - FRUIT=apple
  - VEGETABLE=carrot
`)

	expected := `
apiVersion: v1
data:
  FRUIT: apple
  VEGETABLE: carrot
kind: ConfigMap
metadata:
  name: mymap-kfd8tf729k
`

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}
