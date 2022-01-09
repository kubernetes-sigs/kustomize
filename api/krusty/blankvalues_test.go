package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// test for https://github.com/kubernetes-sigs/kustomize/issues/4240
func TestBlankNamespace4240(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- resource.yaml
`)

	th.WriteF("resource.yaml", `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: test
rules: []
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: test
subjects:
- kind: ServiceAccount
  name: test
  namespace:
roleRef:
  kind: Role
  name: test
  apiGroup: rbac.authorization.k8s.io	
`)

	err := th.RunWithErr(".", th.MakeDefaultOptions())
	if !strings.Contains(err.Error(), "Invalid Input: namespace is blank for resource") {
		t.Fatalf("unexpected error: %q", err)
	}
}
