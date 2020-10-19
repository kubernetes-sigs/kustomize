package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// https://github.com/kubernetes-sigs/kustomize/issues/2640
func TestNameUpdateInRoleRef(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("/app/rbac.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: my-role
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: my-role
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: my-role
subjects:
- kind: ServiceAccount
  name: default
  namespace: foo
--- 
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-role
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-role
roleRef:
  apiGroup: rbac.authorization.k8s.io
  version: v1
  kind: Role
  name: my-role
subjects:
- kind: ServiceAccount
  name: default
`)

	th.WriteK("/app", `
namespace: foo
resources:
- rbac.yaml

patches:
- patch: |-
    - op: add
      path: /metadata/name
      value: prefix_my-role
  target:
    group: rbac.authorization.k8s.io
    version: v1
    kind: ClusterRole
    name: my-role
`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prefix_my-role
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: my-role
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prefix_my-role
subjects:
- kind: ServiceAccount
  name: default
  namespace: foo
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-role
  namespace: foo
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-role
  namespace: foo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-role
  version: v1
subjects:
- kind: ServiceAccount
  name: default
  namespace: foo
`)
}

// https://github.com/kubernetes-sigs/kustomize/issues/3073
func TestNameUpdateInRoleRef2(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("/app/workloads.yaml", `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: myapp

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: myapp
rules:
- apiGroups:
  - ""
  resources:
  - nodes/metrics
  verbs:
  - get

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: myapp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: myapp
subjects:
- kind: ServiceAccount
  name: myapp

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: myapp
rules:
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: myapp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: myapp
subjects:
- kind: ServiceAccount
  name: myapp
`)

	th.WriteF("/app/suffixTransformer.yaml", `
apiVersion: builtin
kind: PrefixSuffixTransformer
metadata:
  name: notImportantHere
suffix: -suffix
fieldSpecs:
- path: metadata/name
  kind: ClusterRole
  name: myapp
- path: metadata/name
  kind: ClusterRoleBinding
  name: myapp
`)

	th.WriteK("/app", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- workloads.yaml
transformers:
- suffixTransformer.yaml
namespace: test

`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: myapp
  namespace: test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: myapp-suffix
rules:
- apiGroups:
  - ""
  resources:
  - nodes/metrics
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: myapp-suffix
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: myapp-suffix
subjects:
- kind: ServiceAccount
  name: myapp
  namespace: test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: myapp
  namespace: test
rules:
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: myapp
  namespace: test
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: myapp
subjects:
- kind: ServiceAccount
  name: myapp
  namespace: test
`)
}
