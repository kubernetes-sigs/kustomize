// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"strings"
	"testing"
)

func TestNamespacedSecrets(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app")

	th.WriteF("/app/secrets.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: dummy
  namespace: default
type: Opaque
data:
  dummy: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: dummy
  namespace: kube-system
type: Opaque
data:
  dummy: ""
`)

	// This should find the proper secret.
	th.WriteF("/app/role.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: dummy
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["dummy"]
  verbs: ["get"]
`)

	th.WriteK("/app", `
resources:
- secrets.yaml
- role.yaml
`)

	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	// TODO: Fix #1044
	// This should not be an error -
	// the secrets have the same name but are in different namespaces.
	// The ClusterRole (by def) is not in a namespace,
	// an in this case applies to *any* Secret resource
	// named "dummy"
	if err == nil {
		t.Fatalf("unexpected lack of error")
	}
	if !strings.Contains(
		err.Error(),
		"slice case - multiple matches for ~G_v1_Secret|default|dummy") {
		t.Fatalf("unexpected error: %s", err)
	}
}
