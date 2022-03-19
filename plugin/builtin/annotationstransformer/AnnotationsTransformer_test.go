// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestAnnotationsTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("AnnotationsTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: AnnotationsTransformer
metadata:
  name: notImportantHere
annotations:
  app: myApp
  greeting/morning: a string with blanks
  booleanNaked: true
  booleanQuoted: "true"
  numberNaked: 42
  numberQuoted: "42"
fieldSpecs:
  - path: metadata/annotations
    create: true
`, `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    app: myApp
    booleanNaked: "true"
    booleanQuoted: "true"
    greeting/morning: a string with blanks
    numberNaked: "42"
    numberQuoted: "42"
  name: myService
spec:
  ports:
  - port: 7002
`)
}

func TestAnnotationsTransformerFieldSpecStrategyReplace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("AnnotationsTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: AnnotationsTransformer
metadata:
  name: notImportantHere
annotations:
  app: bar
fieldSpecs:
  - path: spec/template/metadata/annotations
    create: true
fieldSpecStrategy: replace
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  annotations:
    app: foo
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: foo
  name: myDeployment
spec:
  template:
    metadata:
      annotations:
        app: bar
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestAnnotationsTransformerFieldSpecStrategyMergeDefaults(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("AnnotationsTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: AnnotationsTransformer
metadata:
  name: notImportantHere
annotations:
  app: bar
fieldSpecStrategy: merge
  `, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
---
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: bar
  name: myDeployment
spec:
  template:
    metadata:
      annotations:
        app: bar
    spec:
      containers:
      - image: nginx
        name: nginx
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    app: bar
  name: myService
spec:
  ports:
  - port: 7002
`)
}

func TestAnnotationsTransformerFieldSpecStrategyMergeOverrides(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("AnnotationsTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: AnnotationsTransformer
metadata:
  name: notImportantHere
annotations:
  app: bar
fieldSpecs:
- path: spec/invalid/schema
  create: true
fieldSpecStrategy: merge
  `, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx
---
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: bar
  name: myDeployment
spec:
  invalid:
    schema:
      app: bar
  template:
    metadata:
      annotations:
        app: bar
    spec:
      containers:
      - image: nginx
        name: nginx
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    app: bar
  name: myService
spec:
  invalid:
    schema:
      app: bar
  ports:
  - port: 7002
`)
}
