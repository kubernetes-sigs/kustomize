// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
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
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
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
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "cannot read file \"/app/jsonpatch.json\"") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBadPatchJson6902Transformer(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
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
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "cannot unmarshal string into Go value of type jsonpatch") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBothEmptyJson6902Transformer(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "empty file path and empty jsonOp") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBothSpecifiedJson6902Transformer(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/jsonpatch.json", `[
{"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
{"op": "add", "path": "/spec/replica", "value": "999"},
{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["arg1", "arg2", "arg3"]}
]`)

	_, err := th.RunTransformer(`
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
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "must specify a file path or jsonOp, not both") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPatchJson6902TransformerFromJsonFile(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/jsonpatch.json", `[
{"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
{"op": "add", "path": "/spec/replica", "value": "999"},
{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["arg1", "arg2", "arg3"]}
]`)

	rm := th.LoadAndRunTransformer(`
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
`, target)

	th.AssertActualEqualsExpected(rm, `
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
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/jsonpatch.json", `
- op: add
  path: /spec/template/spec/containers/0/command
  value: ["arg1", "arg2", "arg3"]
`)

	rm := th.LoadAndRunTransformer(`
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
`, target)

	th.AssertActualEqualsExpected(rm, `
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
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
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
`, target)

	th.AssertActualEqualsExpected(rm, `
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
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
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
`, target)

	th.AssertActualEqualsExpected(rm, `
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
