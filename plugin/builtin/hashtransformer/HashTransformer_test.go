// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

func TestHashTransformer(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "HashTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

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

	th.AssertActualEqualsExpected(rm, `
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
