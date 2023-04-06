// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package starlarkmixer_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestStarlarkMixerPlugin(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	rm := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: StarlarkMixer
metadata:
  name: mixer-instance
  annotations:
    config.kubernetes.io/function: |
      starlark:
        path: mixer.star
`,
		`apiVersion: v1
kind: ConfigMap
metadata:
  name: some-cm
data:
  foo: bar
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: delete-me
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    modified-by: mixer-instance
  name: some-cm
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: some-cm-copy
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: net-new
`)
}

func TestStarlarkMixerPlugin_duplicate_rejection(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: someteam.example.com/v1
kind: StarlarkMixer
metadata:
  name: mixer-instance
  annotations:
    config.kubernetes.io/function: |
      starlark:
        path: mixer.star
`,
		`apiVersion: v1
kind: ConfigMap
metadata:
  name: some-cm
data:
  foo: bar
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: some-cm-copy
data:
  foo: bar
`,
		func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error")
			}
			expectedErr := "plugin api: someteam.example.com/v1, " +
				"kind: StarlarkMixer, " +
				"name: mixer-instance generated duplicate " +
				"resource: {\"version\":\"v1\",\"kind\":\"ConfigMap\",\"name\":\"some-cm-copy\"}"
			if !strings.Contains(err.Error(), expectedErr) {
				t.Fatalf("unexpected error: %s", err.Error())
			}
		},
	)
}
