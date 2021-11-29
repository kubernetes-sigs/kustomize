// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestSuffixTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("SuffixTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: SuffixTransformer
metadata:
  name: notImportantHere
suffix: -pie
fieldSpecs:
  - path: metadata/name
`, `
apiVersion: v1
kind: Namespace
metadata:
  name: apple
---
apiVersion: v1
kind: Service
metadata:
  name: apple
spec:
  ports:
  - port: 7002
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crd
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: apiservice
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: Namespace
metadata:
  name: apple
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: Service
    internal.config.kubernetes.io/previousNames: apple
    internal.config.kubernetes.io/previousNamespaces: default
    internal.config.kubernetes.io/suffixes: -pie
  name: apple-pie
spec:
  ports:
  - port: 7002
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crd
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: apiservice
---
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: ConfigMap
    internal.config.kubernetes.io/previousNames: cm
    internal.config.kubernetes.io/previousNamespaces: default
    internal.config.kubernetes.io/suffixes: -pie
  name: cm-pie
`)

	rm = th.LoadAndRunTransformer(`
apiVersion: builtin
kind: SuffixTransformer
metadata:
  name: notImportantHere
suffix: -test
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
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crd
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: apiservice
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
  annotations:
    internal.config.kubernetes.io/previousKinds: Deployment
    internal.config.kubernetes.io/previousNames: deployment
    internal.config.kubernetes.io/previousNamespaces: default
    internal.config.kubernetes.io/suffixes: -test
  name: deployment-test
spec:
  template:
    spec:
      containers:
      - image: myapp
        name: main-test
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crd
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: apiservice
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
`)
}
