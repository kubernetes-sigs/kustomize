// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func makeCommonFileForMergeEnvFromTest(th kusttest_test.Harness) {
	th.WriteF("/app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: image1
        envFrom:
        - configMapRef:
            name: some-config
        - configMapRef:
            name: more-config
`)
}

// When patching, `envFrom` should merge the list instead of replacing it.
func TestMergeEnvFrom(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForMergeEnvFromTest(th)
	th.WriteK("/app", `
resources:
- deployment.yaml

patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
        - name: nginx
          envFrom:
          - configMapRef:
              name: another-config
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: another-config
        image: image1
        name: nginx
`)
}

func TestMergeEnvFromViaJsonInline(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeCommonFileForMergeEnvFromTest(th)
	th.WriteK("app", `
resources:
- deployment.yaml
patches:
- target:
    kind: Deployment
    name: nginx
  patch: |-
    - op: add
      path: /spec/template/spec/containers/0/envFrom/-
      value:
        configMapRef:
          name: another-config
`)
	m := th.Run("app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: some-config
        - configMapRef:
            name: more-config
        - configMapRef:
            name: another-config
        image: image1
        name: nginx
`)
}
