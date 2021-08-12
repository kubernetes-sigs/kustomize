// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSecretGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("SecretGenerator")
	defer th.Reset()

	th.WriteF("a.env", `
ROUTER_PASSWORD=admin
`)
	th.WriteF("b.env", `
DB_PASSWORD=iloveyou
`)
	th.WriteF("longsecret.txt", `
Lorem ipsum dolor sit amet,
consectetur adipiscing elit.
`)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: SecretGenerator
metadata:
  name: mySecret
  namespace: whatever
behavior: merge
envs:
- a.env
- b.env
files:
- obscure=longsecret.txt
literals:
- FRUIT=apple
- VEGETABLE=carrot
- v2=000000000111111111122222222223333333333444444444455
- v3=0000000001111111111222222222233333333334444444444555
`)
	th.AssertActualEqualsExpected(
		rm, `
apiVersion: v1
data:
  DB_PASSWORD: aWxvdmV5b3U=
  FRUIT: YXBwbGU=
  ROUTER_PASSWORD: YWRtaW4=
  VEGETABLE: Y2Fycm90
  obscure: |
    CkxvcmVtIGlwc3VtIGRvbG9yIHNpdCBhbWV0LApjb25zZWN0ZXR1ciBhZGlwaXNjaW5nIG
    VsaXQuCg==
  v2: MDAwMDAwMDAwMTExMTExMTExMTIyMjIyMjIyMjIzMzMzMzMzMzMzNDQ0NDQ0NDQ0NDU1
  v3: MDAwMDAwMDAwMTExMTExMTExMTIyMjIyMjIyMjIzMzMzMzMzMzMzNDQ0NDQ0NDQ0NDU1NQ==
kind: Secret
metadata:
  name: mySecret
  namespace: whatever
type: Opaque
`)
}
