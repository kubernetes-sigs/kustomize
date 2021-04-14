// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestReplacementTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ReplacementTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source: 
    kind: Deployment
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      kind: Deployment
    fieldPaths: 
    - spec.template.spec.containers.1.image
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: foobar:1
        name: replaced-with-digest
      - image: foobar:1
        name: postgresdb
`)
}

func TestReplacementTransformerFromPath(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ReplacementTransformer")
	defer th.Reset()

	th.WriteF("replacement.yaml", `
source: 
  kind: Deployment
  fieldPath: spec.template.spec.containers.0.image
targets:
- select:
    kind: Deployment
  fieldPaths: 
  - spec.template.spec.containers.1.image`)

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- path: replacement.yaml
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: foobar:1
        name: replaced-with-digest
      - image: foobar:1
        name: postgresdb
`)
}

func TestReplacementTransformerComplexType(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ReplacementTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    kind: Pod
    name: pod
    fieldPath: spec.containers
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers
`, `
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers: {}
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
`)
}
