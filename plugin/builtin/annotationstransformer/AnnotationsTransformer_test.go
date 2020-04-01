// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"fmt"
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
yamlSupport: %v
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
	for _, b := range []bool{true, false} {
		t.Run(fmt.Sprintf("yaml-%v", b), func(t *testing.T) {
			th := kusttest_test.MakeEnhancedHarness(t).
				PrepBuiltin("AnnotationsTransformer")
			defer th.Reset()

			cfg := fmt.Sprintf(config, b)
			rm := th.LoadAndRunTransformer(cfg, input)

			th.AssertActualEqualsExpected(rm, expectedOutput)
		})
	}
}
