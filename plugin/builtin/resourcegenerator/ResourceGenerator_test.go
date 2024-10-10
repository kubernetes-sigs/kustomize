// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestResourceGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ResourceGenerator")
	defer th.Reset()

	th.WriteF("config.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: myMap
data:
  COLOR: red
  FRUIT: apple
`)
	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: ResourceGenerator
metadata:
  name: myMap
resource: config.yaml
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: myMap
data:
  COLOR: red
  FRUIT: apple
`)
}
