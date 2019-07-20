// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
)

func writeBase(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/base", `
resources:
- serviceaccount.yaml
- rolebinding.yaml
- clusterrolebinding.yaml
- clusterrole.yaml
namePrefix: pfx-
nameSuffix: -sfx
`)
	th.WriteF("/app/base/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: serviceaccount
`)
	th.WriteF("/app/base/rolebinding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: role
subjects:
- kind: ServiceAccount
  name: serviceaccount
`)
	th.WriteF("/app/base/clusterrolebinding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: role
subjects:
- kind: ServiceAccount
  name: serviceaccount
`)
	th.WriteF("/app/base/clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: role
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "watch", "list"]
`)
}

func writeMidOverlays(th *kusttest_test.KustTestHarness) {
	// Mid-level overlays
	th.WriteK("/app/overlays/a", `
resources:
- ../../base
namePrefix: a-
nameSuffix: -suffixA
`)
	th.WriteK("/app/overlays/b", `
resources:
- ../../base
namePrefix: b-
nameSuffix: -suffixB
`)
}

func writeTopOverlay(th *kusttest_test.KustTestHarness) {
	// Top overlay, combining the mid-level overlays
	th.WriteK("/app/combined", `
resources:
- ../overlays/a
- ../overlays/b
`)
}

func TestBase(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/base")
	writeBase(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pfx-serviceaccount-sfx
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: pfx-rolebinding-sfx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pfx-role-sfx
subjects:
- kind: ServiceAccount
  name: pfx-serviceaccount-sfx
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: pfx-rolebinding-sfx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pfx-role-sfx
subjects:
- kind: ServiceAccount
  name: pfx-serviceaccount-sfx
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pfx-role-sfx
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - watch
  - list
`)
}

func TestMidLevelA(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/overlays/a")
	writeBase(th)
	writeMidOverlays(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: a-pfx-rolebinding-sfx-suffixA
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: a-pfx-role-sfx-suffixA
subjects:
- kind: ServiceAccount
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: a-pfx-rolebinding-sfx-suffixA
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: a-pfx-role-sfx-suffixA
subjects:
- kind: ServiceAccount
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: a-pfx-role-sfx-suffixA
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - watch
  - list
`)
}

func TestMidLevelB(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/overlays/b")
	writeBase(th)
	writeMidOverlays(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: b-pfx-rolebinding-sfx-suffixB
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: b-pfx-role-sfx-suffixB
subjects:
- kind: ServiceAccount
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: b-pfx-rolebinding-sfx-suffixB
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: b-pfx-role-sfx-suffixB
subjects:
- kind: ServiceAccount
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: b-pfx-role-sfx-suffixB
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - watch
  - list
`)
}

func TestMultibasesNoConflict(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/combined")
	writeBase(th)
	writeMidOverlays(th)
	writeTopOverlay(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: a-pfx-rolebinding-sfx-suffixA
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: a-pfx-role-sfx-suffixA
subjects:
- kind: ServiceAccount
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: a-pfx-rolebinding-sfx-suffixA
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: a-pfx-role-sfx-suffixA
subjects:
- kind: ServiceAccount
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: a-pfx-role-sfx-suffixA
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - watch
  - list
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: b-pfx-rolebinding-sfx-suffixB
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: b-pfx-role-sfx-suffixB
subjects:
- kind: ServiceAccount
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: b-pfx-rolebinding-sfx-suffixB
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: b-pfx-role-sfx-suffixB
subjects:
- kind: ServiceAccount
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: b-pfx-role-sfx-suffixB
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - watch
  - list
`)
}

func TestMultibasesWithConflict(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/combined")
	writeBase(th)
	writeMidOverlays(th)
	writeTopOverlay(th)

	th.WriteK("/app/overlays/a", `
namePrefix: a-
nameSuffix: -suffixA
resources:
- serviceaccount.yaml
- ../../base
`)
	// Expect an error because this resource in the overlay
	// matches a resource in the base.
	th.WriteF("/app/overlays/a/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: serviceaccount
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: a-serviceaccount-suffixA
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: a-pfx-rolebinding-sfx-suffixA
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: a-pfx-role-sfx-suffixA
subjects:
- kind: ServiceAccount
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: a-pfx-rolebinding-sfx-suffixA
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: a-pfx-role-sfx-suffixA
subjects:
- kind: ServiceAccount
  name: a-pfx-serviceaccount-sfx-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: a-pfx-role-sfx-suffixA
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - watch
  - list
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: b-pfx-rolebinding-sfx-suffixB
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: b-pfx-role-sfx-suffixB
subjects:
- kind: ServiceAccount
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: b-pfx-rolebinding-sfx-suffixB
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: b-pfx-role-sfx-suffixB
subjects:
- kind: ServiceAccount
  name: b-pfx-serviceaccount-sfx-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: b-pfx-role-sfx-suffixB
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - watch
  - list
`)
}
