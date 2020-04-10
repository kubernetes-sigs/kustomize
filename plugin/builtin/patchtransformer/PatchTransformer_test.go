// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
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
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.yaml
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"'/patch.yaml' doesn't exist") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerBadPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: "thisIsNotAPatch"
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"unable to get either a Strategic Merge Patch or JSON patch 6902 from") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerMissingSelector(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: '[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]'
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"must specify a target for patch") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerBothEmptyPathAndPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "must specify one of patch and path in") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerBothNonEmptyPathAndPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
Path: patch.yaml
Patch: "something"
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "patch and path can't be set at the same time") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerFromFiles(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.WriteF("patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 3
`)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.yaml
target:
  name: .*Deploy
`,
		target,
		`
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
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: '[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "nginx:latest"}]'
target:
  name: .*Deploy
  kind: Deployment
`, target,
		`
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
