// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

const (
	target = `
apiVersion: apps/v1
metadata:
  name: myDeploy
  labels:
    old-label: old-value
kind: Deployment
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
---
apiVersion: apps/v1
metadata:
  name: yourDeploy
  labels:
    new-label: new-value
kind: Deployment
spec:
  replica: 1
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
---
apiVersion: apps/v1
metadata:
  name: myDeploy
  label:
    old-label: old-value
kind: MyKind
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`
)

func TestPatchTransformerMissingFile(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchTransformer")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.yaml
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(),
		"cannot read file \"/app/patch.yaml\"") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPatchTransformerBadPatch(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchTransformer")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: "thisIsNotAPatch"
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(),
		"unable to get either a Strategic Merge Patch or JSON patch 6902 from") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPatchTransformerMissingSelector(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchTransformer")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: '[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]'
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(),
		"must specify a target for patch") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPatchTransformerBothEmptyPathAndPatch(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "must specify one of patch and path in") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPatchTransformerBothNonEmptyPathAndPatch(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
Path: patch.yaml
Patch: "something"
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "patch and path can't be set at the same time") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPatchTransformerFromFiles(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 3
`)

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.yaml
target:
  name: .*Deploy
`, target)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    old-label: old-value
  name: myDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchTransformerWithInline(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: '[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "nginx:latest"}]'
target:
  name: .*Deploy
  kind: Deployment
`, target)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    old-label: old-value
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx:latest
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 1
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - image: nginx:latest
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}
