// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// See https://github.com/kubernetes-sigs/kustomize/issues/1878
// User provided incorrect kustomization.yaml
// Kustomize could do a better job at detecting/warning regarding this
func TestManyPatchesInYAML(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteK("/app/base", `
resources:
- pod.yaml
patches:
- path: patch1.yaml
  target:
	kind: Pod
patches:
- path: patch2.yaml
  target:
	kind: Pod
`)

	th.WriteF("/app/base/pod.yaml", `
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
spec:
  containers:
  - name: myapp-container
	image: busybox:1.29.0
	command: ['sleep 3600']
`)

	th.WriteF("/app/base/patch1.yaml", `
- op: add
  path: /spec/labelSelector
  value:
    patch1: was-here
`)

	th.WriteF("/app/base/patch2.yaml", `
- op: add
  path: /spec/nodeSelector
  value:
    patch2: was-here-too
`)

	m := th.Run("/app/base", th.MakeDefaultOptions())

	//nolint
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: PodXX
metadata:
  name: myapp-pod
spec:
  containers:
  - command:
	- sleep 3600
	image: busybox:1.29.0
	name: myapp-container
  labelSelector:
	patch1: was-here
  nodeSelector:
	patch2: was-here-too
`)
}
