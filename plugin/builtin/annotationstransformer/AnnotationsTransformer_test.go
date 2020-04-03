// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

var (
	config = `
apiVersion: builtin
kind: AnnotationsTransformer
metadata:
  name: notImportantHere
annotations:
  app: myApp
  greeting/morning: a string with blanks
fieldSpecs:
  - path: metadata/annotations
    create: true
`
	input = `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`
	expectedOutput = `
apiVersion: v1
kind: Service
metadata:
  annotations:
    app: myApp
    greeting/morning: a string with blanks
  name: myService
spec:
  ports:
  - port: 7002
`
)

func TestAnnotationsTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("AnnotationsTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(config, input, expectedOutput)
}
