package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestMultibyteCharInComment(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
  - resources.yaml
patchesStrategicMerge:
  - patch.yaml
`)
	th.WriteF("resources.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: nginx
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx  # あ
`)
	th.WriteF("patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: nginx
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: my-nginx # あ
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: nginx
spec:
  template:
    spec:
      containers:
      - image: my-nginx
        name: nginx
`)
}
