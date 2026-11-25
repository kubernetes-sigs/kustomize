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

	m := th.Run(".", th.MakeDefaultOptions())
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
  namespace: b
`)
}

// TestClusterRoleBindingSubjectNamespaceFromSharedBase is a regression test
// for https://github.com/kubernetes-sigs/kustomize/issues/3877.
// One ServiceAccount base is included from two overlays that set different
// namespaces. Each ClusterRoleBinding must refer to the ServiceAccount copy
// from its own overlay.
func TestClusterRoleBindingSubjectNamespaceFromSharedBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("base/resources.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
`)
	th.WriteK("base", `
resources:
- resources.yaml
`)

	th.WriteF("a/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-a
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK("a", `
resources:
- ../base
- resources.yaml
namespace: a
`)

	th.WriteF("b/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-b
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK("b", `
resources:
- ../base
- resources.yaml
namespace: b
`)

	th.WriteK(".", `
resources:
- a
- b
`)

	m := th.Run(".", th.MakeDefaultOptions())
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
  name: cr
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
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
  namespace: b
`)
}

// The ServiceAccount is namespaced in the shared base, then again in each
// overlay. The binding is only in the overlay, so its recorded namespace
// context is a suffix of the ServiceAccount's context.
func TestClusterRoleBindingSubjectNamespaceWhenBaseAlsoSetsNamespace(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("base/resources.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
`)
	th.WriteK("base", `
namespace: shared
resources:
- resources.yaml
`)

	th.WriteF("a/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-a
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK("a", `
namespace: a
resources:
- ../base
- resources.yaml
`)

	th.WriteF("b/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-b
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK("b", `
namespace: b
resources:
- ../base
- resources.yaml
`)

	th.WriteK(".", `
resources:
- a
- b
`)

	m := th.Run(".", th.MakeDefaultOptions())
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
  name: cr
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
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
  namespace: b
`)
}

func TestClusterRoleBindingSubjectNamespaceWithNamePrefix(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("base/resources.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
`)
	th.WriteK("base", `
resources:
- resources.yaml
`)

	th.WriteF("a/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK("a", `
namePrefix: a-
namespace: a
resources:
- ../base
- resources.yaml
`)

	th.WriteF("b/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK("b", `
namePrefix: b-
namespace: b
resources:
- ../base
- resources.yaml
`)

	th.WriteK(".", `
resources:
- a
- b
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: a-sa
  namespace: a
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: a-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: a-sa
  namespace: a
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: b-sa
  namespace: b
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: b-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: b-sa
  namespace: b
`)
}

func TestRoleBindingSubjectNamespaceFromSharedBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("base/resources.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
`)
	th.WriteK("base", `
resources:
- resources.yaml
`)

	th.WriteF("a/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rb-a
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK("a", `
namespace: a
resources:
- ../base
- resources.yaml
`)

	th.WriteF("b/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rb-b
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK("b", `
namespace: b
resources:
- ../base
- resources.yaml
`)

	th.WriteK(".", `
resources:
- a
- b
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
  namespace: a
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rb-a
  namespace: a
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
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
kind: RoleBinding
metadata:
  name: rb-b
  namespace: b
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
  namespace: b
`)
}

// A ClusterRoleBinding written at the root has no namespace context of its
// own, so two ServiceAccounts with the same name stay ambiguous.
func TestClusterRoleBindingSubjectNamespaceAmbiguousAtRoot(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("a/sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
`)
	th.WriteK("a", `
namespace: a
resources:
- sa.yaml
`)

	th.WriteF("b/sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
`)
	th.WriteK("b", `
namespace: b
resources:
- sa.yaml
`)

	th.WriteF("crb.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
`)
	th.WriteK(".", `
resources:
- a
- b
- crb.yaml
`)

	err := th.RunWithErr(".", th.MakeDefaultOptions())
	assert.ErrorContains(t, err, "found multiple possible referrals")
	assert.ErrorContains(t, err, "ServiceAccount.v1.[noGrp]/sa.a")
	assert.ErrorContains(t, err, "ServiceAccount.v1.[noGrp]/sa.b")
}

// An explicit subject namespace still selects that namespace when another
// ServiceAccount with the same name exists.
func TestClusterRoleBindingExplicitSubjectNamespace(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("a/resources.yaml", `
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
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
  namespace: b
`)
	th.WriteK("a", `
namespace: a
resources:
- resources.yaml
`)

	th.WriteF("b/resources.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
`)
	th.WriteK("b", `
namespace: b
resources:
- resources.yaml
`)

	th.WriteK(".", `
resources:
- a
- b
`)

	m := th.Run(".", th.MakeDefaultOptions())
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
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
  namespace: b
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
  namespace: b
`)
}

// Comment on https://github.com/kubernetes-sigs/kustomize/issues/3877:
// a nameSuffix and the kustomization namespace were not applied to a
// ClusterRoleBinding subject, and a subject namespace of "default" stayed
// "default".
func TestClusterRoleBindingSubjectGetsNameSuffixAndNamespace(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("resources.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
  namespace: default
`)
	th.WriteK(".", `
nameSuffix: -sfx
namespace: prod
resources:
- resources.yaml
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-sfx
  namespace: prod
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-sfx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa-sfx
  namespace: prod
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb-default-sfx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa-sfx
  namespace: prod
`)
}

// Same report, with the shared base from #3877: each overlay's suffix and
// namespace must be written onto a subject that was explicitly "default".
func TestClusterRoleBindingDefaultSubjectNamespaceAcrossOverlays(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/resources.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa
`)
	th.WriteK("base", `
resources:
- resources.yaml
`)

	th.WriteF("a/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
  namespace: default
`)
	th.WriteK("a", `
nameSuffix: -a
namespace: a
resources:
- ../base
- resources.yaml
`)

	th.WriteF("b/resources.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cr
subjects:
- kind: ServiceAccount
  name: sa
  namespace: default
`)
	th.WriteK("b", `
nameSuffix: -b
namespace: b
resources:
- ../base
- resources.yaml
`)

	th.WriteK(".", `
resources:
- a
- b
`)

	m := th.Run(".", th.MakeDefaultOptions())
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
  name: cr
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
  name: cr
subjects:
- kind: ServiceAccount
  name: sa-b
  namespace: b
`)
}
