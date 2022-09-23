// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestHashTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("HashTransformer")
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: HashTransformer
metadata:
  name: myMapGen
name: myMap
envs:
- devops.env
- uxteam.env
literals:
- FRUIT=apple
- VEGETABLE=carrot
`, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: Secret
metadata:
  name: secret1
---
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-lab: old-val
    spec:
      containers:
      - name: ngnix
        image: nginx:1.7.9
`)

	th.AssertActualEqualsExpectedNoIdAnnotations(rm, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: Secret
metadata:
  name: secret1
---
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-lab: old-val
    spec:
      containers:
      - image: nginx:1.7.9
        name: ngnix
`)
}
