// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

// Demo custom configuration of a builtin transformation.
// This is a NamePrefixer that only touches Deployments
// and Services.
func TestCustomNamePrefixer(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PrefixSuffixTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteK("/app", `
resources:
- deployment.yaml
- role.yaml
- service.yaml
transformers:
- prefixer.yaml
`)
	th.WriteF("/app/prefixer.yaml", `
apiVersion: builtin
kind: PrefixSuffixTransformer
metadata:
  name: customPrefixer
prefix: zzz-
fieldSpecs:
- kind: Deployment
  path: metadata/name
- kind: Service
  path: metadata/name
`)
	th.WriteF("/app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - name: whatever
        image: whatever
`)
	th.WriteF("/app/role.yaml", `
apiVersion: v1
kind: Role
metadata:
  name: myRole
`)
	th.WriteF("/app/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zzz-myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
---
apiVersion: v1
kind: Role
metadata:
  name: myRole
---
apiVersion: v1
kind: Service
metadata:
  name: zzz-myService
`)
}
