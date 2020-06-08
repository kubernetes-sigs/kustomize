// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestAddManagedbyLabel(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("/app/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`)
	th.WriteK("/app", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
`)
	options := th.MakeDefaultOptions()
	options.AddManagedbyLabel = true
	m := th.Run("/app", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/managed-by: kustomize-v444.333.222
  name: myService
spec:
  ports:
  - port: 7002
`)
}
