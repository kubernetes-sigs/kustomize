// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package starlarkmixer_test

import (
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
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/path: configmap_some-cm.yaml
    modified-by: mixer-instance
  name: some-cm
---
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/path: configmap_some-cm-copy.yaml
  name: some-cm-copy
---
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/path: configmap_net-new.yaml
  name: net-new
`)
}
