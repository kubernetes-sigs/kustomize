// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestNamedspacedServiceAccountsWithoutOverlap(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("a/a.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-a
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
    name: sa-a
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
  name: sa-b
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
    name: sa-b
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

	m := th.Run(".", th.MakeDefaultOptions())

	// Everything is as expected: each CRB gets updated to reference the SA in the appropriate namespace
	// Changing order in the root kustomization.yaml does not change the result, as expected
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-a
  namespace: a
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
  name: sa-a
  namespace: a
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-b
  namespace: b
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
  name: sa-b
  namespace: b
`)
}

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

	m := th.Run(".", th.MakeDefaultOptions())

	// Unexpected result: crb-b's subject obtains the wrong namespace, having "namespace: a"
	// If the order is swapped in the kustomization.yaml then it's crb-a that gets "namespace: b"
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
  namespace: a
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
  namespace: a
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
  namespace: b
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
  namespace: a
`)
}
