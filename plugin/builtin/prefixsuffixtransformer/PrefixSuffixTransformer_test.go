// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

func TestPrefixSuffixTransformer(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PrefixSuffixTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")
	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PrefixSuffixTransformer
metadata:
  name: notImportantHere
prefix: baked-
suffix: -pie
fieldSpecs:
  - path: metadata/name
`, `
apiVersion: v1
kind: Service
metadata:
  name: apple
spec:
  ports:
  - port: 7002
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crd
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: Service
metadata:
  name: baked-apple-pie
spec:
  ports:
  - port: 7002
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crd
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: baked-cm-pie
`)

	rm = th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PrefixSuffixTransformer
metadata:
  name: notImportantHere
prefix: test-
fieldSpecs:
  - kind: Deployment
    path: metadata/name
  - kind: Deployment
    path: spec/template/spec/containers/name
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - image: myapp
        name: main
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crd
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: myapp
        name: test-main
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crd
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
`)
}
