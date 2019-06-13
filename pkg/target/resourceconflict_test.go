// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/pkg/kusttest"
)

func writeCombinedOverlays(th *kusttest_test.KustTestHarness) {
	// Base
	th.WriteK("/app/base", `
resources:
- serviceaccount.yaml
- rolebinding.yaml
namePrefix: base-
nameSuffix: -suffix
`)
	th.WriteF("/app/base/rolebinding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: role
subjects:
- kind: ServiceAccount
  name: serviceaccount
`)
	th.WriteF("/app/base/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: serviceaccount
`)

	// Mid-level overlays
	th.WriteK("/app/overlays/a", `
bases:
- ../../base
namePrefix: a-
nameSuffix: -suffixA
`)
	th.WriteK("/app/overlays/b", `
bases:
- ../../base
namePrefix: b-
nameSuffix: -suffixB
`)

	// Top overlay, combining the mid-level overlays
	th.WriteK("/app/combined", `
bases:
- ../overlays/a
- ../overlays/b
`)
}

func TestMultibasesNoConflict(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/combined")
	writeCombinedOverlays(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: a-base-serviceaccount-suffix-suffixA
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: b-base-serviceaccount-suffix-suffixB
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: a-base-rolebinding-suffix-suffixA
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: role
subjects:
- kind: ServiceAccount
  name: a-base-serviceaccount-suffix-suffixA
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: b-base-rolebinding-suffix-suffixB
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: role
subjects:
- kind: ServiceAccount
  name: b-base-serviceaccount-suffix-suffixB
`)
}

func TestMultibasesWithConflict(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/combined")
	writeCombinedOverlays(th)

	th.WriteK("/app/overlays/a", `
bases:
- ../../base
namePrefix: a-
nameSuffix: -suffixA
resources:
- serviceaccount.yaml
`)
	// Expect an error because this resource in the overlay
	// matches a resource in the base.
	th.WriteF("/app/overlays/a/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: serviceaccount
`)

	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("Expected resource conflict.")
	}
	if !strings.Contains(
		err.Error(), "Multiple matches for name ~G_v1_ServiceAccount") {
		t.Fatalf("Unexpected err: %v", err)
	}
}
