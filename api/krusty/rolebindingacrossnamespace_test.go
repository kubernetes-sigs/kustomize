package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestRoleBindingAcrossNamespace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("/app", `
resources:
- resource.yaml
nameSuffix: -ns2
`)
	th.WriteF("/app/resource.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa1
  namespace: ns1
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa2
  namespace: ns2
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa3
  namespace: ns3
---
apiVersion: v1
kind: NotServiceAccount
metadata:
  name: my-nsa
  namespace: ns1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-role
  namespace: ns2
rules:
  - apiGroups:
      - '*'
    resources:
      - '*'
    verbs:
      - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-role-binding
  namespace: ns2
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-role
subjects:
  - kind: ServiceAccount
    name: my-sa1
    namespace: ns1
  - kind: ServiceAccount
    name: my-sa2
    namespace: ns2
  - kind: ServiceAccount
    name: my-sa3
    namespace: ns3
  - kind: NotServiceAccount
    name: my-nsa
    namespace: ns1
`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa1-ns2
  namespace: ns1
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa2-ns2
  namespace: ns2
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa3-ns2
  namespace: ns3
---
apiVersion: v1
kind: NotServiceAccount
metadata:
  name: my-nsa-ns2
  namespace: ns1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-role-ns2
  namespace: ns2
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-role-binding-ns2
  namespace: ns2
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-role-ns2
subjects:
- kind: ServiceAccount
  name: my-sa1-ns2
  namespace: ns1
- kind: ServiceAccount
  name: my-sa2-ns2
  namespace: ns2
- kind: ServiceAccount
  name: my-sa3-ns2
  namespace: ns3
- kind: NotServiceAccount
  name: my-nsa
  namespace: ns1
`)
}

func TestRoleBindingAcrossNamespaceWoSubjects(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("/app", `
resources:
- resource.yaml
nameSuffix: -ns2
`)
	th.WriteF("/app/resource.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa1
  namespace: ns1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-role
  namespace: ns2
rules:
  - apiGroups:
      - '*'
    resources:
      - '*'
    verbs:
      - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-role-binding
  namespace: ns2
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-role
`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa1-ns2
  namespace: ns1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-role-ns2
  namespace: ns2
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-role-binding-ns2
  namespace: ns2
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-role-ns2
`)
}

// The ServiceAccount in subjects in role binding can be across namespace
// but the roleRef is not. This test is used to cover such case.
func TestRoleBindingWhenSubjectsAcrossNamespace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()
	th.WriteK("/app", `
resources:
- ./ns1
- ./ns2
`)
	th.WriteK("/app/ns1", `
namespace: namespace-1
resources:
- role-ns1.yaml
- rolebinding-ns1.yaml
`)
	th.WriteF("/app/ns1/role-ns1.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: testRole
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get"]
`)
	th.WriteF("/app/ns1/rolebinding-ns1.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: testRoleBinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: testRole
subjects:
  - kind: ServiceAccount
    name: testAccount
    namespace: namespace-2
`)
	th.WriteK("/app/ns2", `
namespace: namespace-2
resources:
- role-ns2.yaml
- rolebinding-ns2.yaml
`)
	th.WriteF("/app/ns2/role-ns2.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: testRole
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get"]
`)
	th.WriteF("/app/ns2/rolebinding-ns2.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: testRoleBinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: testRole
subjects:
  - kind: ServiceAccount
    name: testAccount
    namespace: namespace-1
`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: testRole
  namespace: namespace-1
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: testRoleBinding
  namespace: namespace-1
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: testRole
subjects:
- kind: ServiceAccount
  name: testAccount
  namespace: namespace-2
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: testRole
  namespace: namespace-2
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: testRoleBinding
  namespace: namespace-2
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: testRole
subjects:
- kind: ServiceAccount
  name: testAccount
  namespace: namespace-1
`)
}
