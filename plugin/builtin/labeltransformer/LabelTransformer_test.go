// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestLabelTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("LabelTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: notImportantHere
labels:
  app: myApp
  quotedBoolean: "true"
  quotedFruit: "peach"
  unquotedBoolean: true
  env: production
fieldSpecs:
- path: spec/selector
  create: true
  version: v1
  kind: Service
- path: metadata/labels
  create: true
- path: spec/selector/matchLabels
  create: true
  kind: Deployment
- path: spec/template/metadata/labels
  create: true
  kind: Deployment
`, `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mungebot
  labels:
    app: mungebot
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mungebot
    spec:
      containers:
      - name: nginx
        image: nginx
`)

	th.AssertActualEqualsExpectedNoIdAnnotations(rm, `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
  name: myService
spec:
  ports:
  - port: 7002
  selector:
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
  name: mungebot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myApp
      env: production
      quotedBoolean: "true"
      quotedFruit: peach
      unquotedBoolean: "true"
  template:
    metadata:
      labels:
        app: myApp
        env: production
        quotedBoolean: "true"
        quotedFruit: peach
        unquotedBoolean: "true"
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestLabelTransformerFieldSpecStrategyReplace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("LabelTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: notImportantHere
labels:
  app: bar
fieldSpecs:
- path: spec/template/metadata/labels
  create: true
  kind: Deployment
fieldSpecStrategy: replace
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mungebot
  labels:
    app: foo
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: foo
    spec:
      containers:
      - name: nginx
        image: nginx
`)

	th.AssertActualEqualsExpectedNoIdAnnotations(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: foo
  name: mungebot
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: bar
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestLabelTransformerFieldSpecStrategyMergeDefaults(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("LabelTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: notImportantHere
labels:
  app: bar
fieldSpecStrategy: merge
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mungebot
  labels:
    app: foo
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: foo
    spec:
      containers:
      - name: nginx
        image: nginx
`)

	th.AssertActualEqualsExpectedNoIdAnnotations(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: bar
  name: mungebot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bar
  template:
    metadata:
      labels:
        app: bar
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestLabelTransformerFieldSpecStrategyMergeOverrides(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("LabelTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: notImportantHere
labels:
  app: bar
fieldSpecs:
- path: spec/invalid/schema
  create: true
  kind: Deployment
fieldSpecStrategy: merge
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mungebot
  labels:
    app: foo
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: foo
    spec:
      containers:
      - name: nginx
        image: nginx
`)

	th.AssertActualEqualsExpectedNoIdAnnotations(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: bar
  name: mungebot
spec:
  invalid:
    schema:
      app: bar
  replicas: 1
  selector:
    matchLabels:
      app: bar
  template:
    metadata:
      labels:
        app: bar
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}
