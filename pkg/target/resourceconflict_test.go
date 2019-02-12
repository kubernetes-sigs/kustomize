/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package target_test

import (
	"strings"
	"testing"
)

func writeCombinedOverlays(th *KustTestHarness) {
	// Base
	th.writeK("/app/base", `
resources:
- serviceaccount.yaml
- rolebinding.yaml
namePrefix: base-
nameSuffix: -suffix
`)
	th.writeF("/app/base/rolebinding.yaml", `
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
	th.writeF("/app/base/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: serviceaccount
`)

	// Mid-level overlays
	th.writeK("/app/overlays/a", `
bases:
- ../../base
namePrefix: a-
nameSuffix: -suffixA
`)
	th.writeK("/app/overlays/b", `
bases:
- ../../base
namePrefix: b-
nameSuffix: -suffixB
`)

	// Top overlay, combining the mid-level overlays
	th.writeK("/app/combined", `
bases:
- ../overlays/a
- ../overlays/b
`)
}

func TestMultibasesNoConflict(t *testing.T) {
	th := NewKustTestHarness(t, "/app/combined")
	writeCombinedOverlays(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
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
	th := NewKustTestHarness(t, "/app/combined")
	writeCombinedOverlays(th)

	th.writeK("/app/overlays/a", `
bases:
- ../../base
namePrefix: a-
nameSuffix: -suffixA
resources:
- serviceaccount.yaml
`)
	// Expect an error because this resource in the overlay
	// matches a resource in the base.
	th.writeF("/app/overlays/a/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: serviceaccount
`)

	_, err := th.makeKustTarget().MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("Expected resource conflict.")
	}
	if !strings.Contains(
		err.Error(), "Multiple matches for name ~G_v1_ServiceAccount") {
		t.Fatalf("Unexpected err: %v", err)
	}
}
