// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestNamedspacedServiceAccountsWithOverlap(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("a/a.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-a
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr-a
subjects:
  - kind: ServiceAccount
    name: sa
`)

	th.WriteK("a/", `
namespace: a
resources:
- a.yaml
`)

	th.WriteF("b/b.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-b
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr-b
subjects:
  - kind: ServiceAccount
    name: sa
`)

	th.WriteK("b/", `
namespace: b
resources:
- b.yaml
`)

	th.WriteK(".", `
resources:
- a
- b
`)

	err := th.RunWithErr(".", th.MakeDefaultOptions())
	assert.Contains(t, err.Error(), "found multiple possible referrals: ServiceAccount.v1.[noGrp]/sa.a, ServiceAccount.v1.[noGrp]/sa.b")
}
