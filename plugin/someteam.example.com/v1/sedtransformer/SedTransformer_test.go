// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSedTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "SedTransformer")
	defer th.Reset()

	th.WriteF("sed-input.txt", `
s/$FRUIT/orange/g
s/$VEGGIE/tomato/g
`)

	rm := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: SedTransformer
metadata:
  name: notImportantHere
argsOneLiner: s/one/two/g
argsFromFile: sed-input.txt
`,
		`apiVersion: apps/v1
kind: MeatBall
metadata:
  name: notImportantHere
beans: one one one one
fruit: $FRUIT
vegetable: $VEGGIE
`)

	th.AssertActualEqualsExpectedNoIdAnnotations(rm, `
apiVersion: apps/v1
beans: two two two two
fruit: orange
kind: MeatBall
metadata:
  name: notImportantHere
vegetable: tomato
`)
}
