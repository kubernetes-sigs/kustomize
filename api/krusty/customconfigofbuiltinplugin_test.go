// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Demo custom configuration of a builtin transformation.
// This is a NamePrefixer that touches Deployments
// and Services exclusively.
func TestCustomNamePrefixer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PrefixSuffixTransformer")
	defer th.Reset()

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

	m := th.Run("/app", th.MakeDefaultOptions())
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
