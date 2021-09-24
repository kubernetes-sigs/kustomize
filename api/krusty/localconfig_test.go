// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// This test checks that if a resource is annotated as "local-config", this resource won't be ignored until
// all transformations have completed. This makes sure the local resource can be used as a transformation input.
// See https://github.com/kubernetes-sigs/kustomize/issues/4124 for details.
func TestSKipLocalConfigAfterTransform(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - pod.yaml
  - deployment.yaml
transformers:
  - replacement.yaml
`)
	th.WriteF("pod.yaml", `apiVersion: v1
kind: Pod
metadata:
  name: buildup
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  containers:
    - name: app
      image: nginx
`)
	th.WriteF("deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: buildup
`)
	th.WriteF("replacement.yaml", `
apiVersion: builtin
kind: ReplacementTransformer
metadata:
  name: buildup
replacements:
- source:
    kind: Pod
    fieldPath: spec
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec
    options:
      create: true
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: buildup
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: app
`)
}
