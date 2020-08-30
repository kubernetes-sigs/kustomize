// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const badTarget = `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
`
const target = `
apiVersion: apps/v1
metadata:
  name: myConfig
kind: ConfigMap
data:
  one-key: one-value
  second-key.json: |-
        {
          "one-json-key": "one-json-value"
        }
`

func TestConfigMapPatchJsonTransformerNotConfigMap(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
`, badTarget, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "This patch cannot be applied to non configmap resources") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestConfigMapPatchJsonTransformerMissingFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: ConfigMap
  name: myConfig
fields:
  - name: some-key
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

func TestBadConfigMapPatchJsonTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
fields:
  - name: some-key
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

func TestBothEmptyConfigMapPatchJsonTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: ConfigMap
  name: myConfig
fields:
  - name: some-key
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "empty file path and empty jsonOp") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestBothSpecifiedConfigMapPatchJsonTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.WriteF("/app/jsonpatch.json", `[
{"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
{"op": "add", "path": "/spec/replica", "value": "999"},
{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["arg1", "arg2", "arg3"]}
]`)

	th.RunTransformerAndCheckError(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: Deployment
  name: myDeploy
fields:
  - name: some-key
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

func TestInvalidKeyConfigMapPatchJsonTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: ConfigMap
  name: myConfig
fields:
  - name: some-non-existent-key
    jsonOp: '[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]'
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "Non existent key some-non-existent-key in configmap") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestNonJsonValueConfigMapPatchJsonTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: ConfigMap
  name: myConfig
fields:
  - name: one-key
    jsonOp: '[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]'
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "invalid character 'o' looking for beginning of value for key one-key") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestConfigMapPatchJsonTransformerFromJsonFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.WriteF("jsonpatch.json", `[
{"op": "replace", "path": "/one-json-key", "value": "updated-json-value"},
{"op": "add", "path": "/second-json-key", "value": 999},
{"op": "add", "path": "/third-json-key", "value": ["arg1", "arg2", "arg3"]}
]`)

	th.RunTransformerAndCheckResult(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: ConfigMap
  name: myConfig
fields:
  - name: second-key.json
    path: jsonpatch.json
`,
		target,
		`
apiVersion: apps/v1
data:
  one-key: one-value
  second-key.json: |-
    {
      "one-json-key": "updated-json-value",
      "second-json-key": 999,
      "third-json-key": [
        "arg1",
        "arg2",
        "arg3"
      ]
    }
kind: ConfigMap
metadata:
  name: myConfig
`)
}

func TestConfigMapPatchJsonTransformerFromYamlFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.WriteF("jsonpatch.json", `
- op: add
  path: /third-json-key
  value: ["arg1", "arg2", "arg3"]
`)

	th.RunTransformerAndCheckResult(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: ConfigMap
  name: myConfig
fields:
  - name: second-key.json
    path: jsonpatch.json
`,
		target,
		`
apiVersion: apps/v1
data:
  one-key: one-value
  second-key.json: |-
    {
      "one-json-key": "one-json-value",
      "third-json-key": [
        "arg1",
        "arg2",
        "arg3"
      ]
    }
kind: ConfigMap
metadata:
  name: myConfig
`)
}

func TestConfigMapPatchJsonTransformerWithInlineJSON(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: ConfigMap
  name: myConfig
fields:
  - name: second-key.json
    jsonOp: '[{"op": "add", "path": "/second-json-key", "value": 999}]'
`,
		target,
		`
apiVersion: apps/v1
data:
  one-key: one-value
  second-key.json: |-
    {
      "one-json-key": "one-json-value",
      "second-json-key": 999
    }
kind: ConfigMap
metadata:
  name: myConfig
`)
}

func TestConfigMapPatchJsonTransformerWithInlineYAML(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "ConfigMapPatchJsonTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: someteam.example.com/v1
kind: ConfigMapPatchJsonTransformer
metadata:
  name: notImportantHere
target:
  group: apps
  version: v1
  kind: ConfigMap
  name: myConfig
fields:
  - name: second-key.json
    jsonOp: |-
      - op: add
        path: /second-json-key
        value: 999
`,
		target,
		`
apiVersion: apps/v1
data:
  one-key: one-value
  second-key.json: |-
    {
      "one-json-key": "one-json-value",
      "second-json-key": 999
    }
kind: ConfigMap
metadata:
  name: myConfig
`)
}
