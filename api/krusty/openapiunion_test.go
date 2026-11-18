// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestBuiltinOpenAPIUnionSupportsRemovedGVK(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- cronjob.yaml
patches:
- path: patch.yaml
`)
	th.WriteF("cronjob.yaml", `
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: example
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: main
            image: example:v1
          - name: sidecar
            image: sidecar:v1
`)
	th.WriteF("patch.yaml", `
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: example
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: main
            image: example:v2
`)

	result := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(result, `
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: example
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - image: example:v2
            name: main
          - image: sidecar:v1
            name: sidecar
`)
}

func TestBuiltinOpenAPIUnionSupportsCurrentGVK(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- policy.yaml
patches:
- path: patch.yaml
`)
	th.WriteF("policy.yaml", `
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingAdmissionPolicy
metadata:
  name: example
spec:
  matchConditions:
  - name: first
    expression: "true"
  - name: second
    expression: "true"
`)
	th.WriteF("patch.yaml", `
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingAdmissionPolicy
metadata:
  name: example
spec:
  matchConditions:
  - name: first
    expression: "false"
`)

	result := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(result, `
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingAdmissionPolicy
metadata:
  name: example
spec:
  matchConditions:
  - expression: "false"
    name: first
  - expression: "true"
    name: second
`)
}
