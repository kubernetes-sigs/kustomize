// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Demo custom configuration of a builtin transformation.
// This is a NamePrefixer that only touches Deployments
// and Services.
func TestCustomNamePrefixer(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PrefixSuffixTransformer")

	th := kusttest_test.NewKustTestHarnessAllowPlugins(t, "/app")

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

// Demo custom configuration as a base.
func TestReusableCustomNamePrefixer(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PrefixSuffixTransformer")
	tc.BuildGoPlugin(
		"builtin", "", "LabelTransformer")

	th := kusttest_test.NewKustTestHarnessAllowPlugins(t, "/app/foo")

	// This kustomization file contains resources that
	// all happen to be plugin configurations. This makes
	// these plugins all available as part of a base,
	// re-usable in any number of other kustomizations.
	// Just specify the path (or URL) to this base in the
	// "transformers:" field (not the "resources" field).
	th.WriteK("/app/mytransformers", `
resources:
- prefixer.yaml
- labeller.yaml
`)
	th.WriteF("/app/mytransformers/prefixer.yaml", `
apiVersion: builtin
kind: PrefixSuffixTransformer
metadata:
  name: myPrefixer
prefix: zzz-
fieldSpecs:
- kind: Deployment
  path: metadata/name
- kind: Service
  path: metadata/name
`)
	th.WriteF("/app/mytransformers/labeller.yaml", `
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: myLabeller
labels:
  company: acmeCorp
fieldSpecs:
- path: spec/template/metadata/labels
  kind: Deployment
`)

	th.WriteK("/app/foo", `
resources:
- deployment.yaml
- role.yaml
- service.yaml
transformers:
- ../mytransformers
`)
	th.WriteF("/app/foo/deployment.yaml", `
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
	th.WriteF("/app/foo/role.yaml", `
apiVersion: v1
kind: Role
metadata:
  name: myRole
`)
	th.WriteF("/app/foo/service.yaml", `
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
        company: acmeCorp
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
