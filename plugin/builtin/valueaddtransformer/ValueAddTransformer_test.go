// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestValueAddTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).PrepBuiltin("ValueAddTransformer")
	th.WriteF("targets.yaml", `
targets:
- fieldPath: metadata/namespace
- selector:
    kind: IAMPolicyMember
  fieldPath: spec/resourceRef/external
  filePathPosition: 2
`)
	defer th.Reset()
	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: ValueAddTransformer
metadata:
  name: notImportantHere
value: area-51
targetFilePath: targets.yaml
`, `
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: iap-tunnel-red
spec:
  member: user:red@example.com
  resourceRef:
    external: projects/data
    kind: Project
  role: roles/engineer
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: not-important
spec:
  template:
    spec:
      containers:
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
`, `
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: iap-tunnel-red
  namespace: area-51
spec:
  member: user:red@example.com
  resourceRef:
    external: projects/area-51/data
    kind: Project
  role: roles/engineer
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: not-important
  namespace: area-51
spec:
  template:
    spec:
      containers:
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
`)
}
