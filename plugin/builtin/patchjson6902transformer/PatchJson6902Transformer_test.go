// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const target = `
apiVersion: apps/v1
metadata:
  name: myDeploy
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
`

func TestPatchJson6902TransformerMissingFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
path: jsonpatch.json
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "'/jsonpatch.json' doesn't exist") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestBadPatchJson6902Transformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
jsonOp: 'thisIsNotAPatch'
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "cannot unmarshal string into Go value of type jsonpatch") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestBothEmptyJson6902Transformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "empty file path and empty jsonOp") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestBothSpecifiedJson6902Transformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	th.WriteF("/app/jsonpatch.json", `[
{"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
{"op": "add", "path": "/spec/replica", "value": "999"},
{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["arg1", "arg2", "arg3"]}
]`)

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
path: jsonpatch.json
jsonOp: '[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]'
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "must specify a file path or jsonOp, not both") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchJson6902TransformerFromJsonFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	th.WriteF("jsonpatch.json", `[
{"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
{"op": "add", "path": "/spec/replica", "value": "999"},
{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["arg1", "arg2", "arg3"]}
]`)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
path: jsonpatch.json
`,
		target,
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: "999"
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - command:
        - arg1
        - arg2
        - arg3
        image: nginx
        name: my-nginx
`)
}

func TestPatchJson6902TransformerFromYamlFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	th.WriteF("jsonpatch.json", `
- op: add
  path: /spec/template/spec/containers/0/command
  value: ["arg1", "arg2", "arg3"]
`)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
path: jsonpatch.json
`,
		target,
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - command:
        - arg1
        - arg2
        - arg3
        image: nginx
        name: nginx
`)
}

func TestPatchJson6902TransformerWithInlineJSON(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
jsonOp: '[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]'
`,
		target,
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
      dnsPolicy: ClusterFirst
`)
}

func TestPatchJson6902TransformerWithInlineYAML(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
jsonOp: |-
  - op: add
    path: /spec/template/spec/dnsPolicy
    value: ClusterFirst
`,
		target,
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
      dnsPolicy: ClusterFirst
`)
}
