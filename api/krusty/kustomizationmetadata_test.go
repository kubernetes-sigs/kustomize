package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestKustomizationMetadata(t *testing.T) {
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
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
metadata:
  annotations:
    config.kubernetes.io/local-config: "true"
  labels:
    foo: bar
  name: test_kustomization
resources:
- resources.yaml  
`)

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
