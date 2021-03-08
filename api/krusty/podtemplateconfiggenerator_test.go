// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Test that a PodTemplate referencing a generated ConfigMap is generated consistently
func TestPodTemplateWithConfigGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PrefixSuffixTransformer").
		PrepBuiltin("AnnotationsTransformer").
		PrepBuiltin("LabelTransformer")
	defer th.Reset()

	// Simple template for an app that mounts a config map
	th.WriteF("podtemplateconfiggenerator/app.yml", `
apiVersion: v1
kind: PodTemplate
metadata:
  name: tmpl
template:
  spec:
    volumes:
      - name: app-config
        configMap:
          name: foo
    containers:
    - name: app
      image: "my-app"
      volumeMounts:
        - mountPath: /usr/src/config
          name: app-config
`)

	// File for config map. Content doesn't matter.
	th.WriteF("podtemplateconfiggenerator/config.py", `hi`)

	// Generate the PodTemplate and ConfigMap
	th.WriteK("podtemplateconfiggenerator", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - app.yml

configMapGenerator:
  - name: foo
    files:
      - config.py
`)

    // the config map name should be consistent. i.e., foo.metadata.name == template.spec.volumes['app-config'].configMap.name
	m := th.Run("podtemplateconfiggenerator", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: PodTemplate
metadata:
  name: tmpl
template:
  spec:
    containers:
    - image: my-app
      name: app
      volumeMounts:
      - mountPath: /usr/src/config
        name: app-config
    volumes:
    - configMap:
        name: foo-m76bm8cb55
      name: app-config
---
apiVersion: v1
data:
  config.py: hi
kind: ConfigMap
metadata:
  name: foo-m76bm8cb55
`)
}
