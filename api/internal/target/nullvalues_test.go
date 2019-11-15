// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestNullValues(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app")
	th.WriteF("/app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: example
  name: example
spec:
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - args: null
        image: image
        name: example
`)
	th.WriteF("/app/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
`)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: example
  name: example
spec:
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - args: null
        image: image
        name: example
`)
}
