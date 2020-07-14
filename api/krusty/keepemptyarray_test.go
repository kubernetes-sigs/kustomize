package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestKeepEmptyArray(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("/app/resources.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testing123
spec:
  replicas: 1
  selector: null
  template:
    spec:
      containers:
      - name: event
        image: testing123
        imagePullPolicy: IfNotPresent
      imagePullSecrets: []`)

	th.WriteK("/app", `
resources:
- resources.yaml`)

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testing123
spec:
  replicas: 1
  selector: null
  template:
    spec:
      containers:
      - image: testing123
        imagePullPolicy: IfNotPresent
        name: event
      imagePullSecrets: []
`)
}
