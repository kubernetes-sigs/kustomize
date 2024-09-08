// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSecretGeneratorAsTransformer(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("app/base", `
transformers:
- |-
  apiVersion: builtin
  kind: SecretGenerator
  metadata:
    name: foo
  literals:
  - foo=bar
  options:
    disableNameSuffixHash: true
`)
	m := th.Run("app/base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  foo: YmFy
kind: Secret
metadata:
  name: foo
type: Opaque
`)
}
