package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestYamlAnchorsInKustomization(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("resources.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
data:
  key1: value1
  key2: value2
  key3: value3
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
data:
  keya: valuea
  keyb: valueb
  keyc: valuec
`)
	th.WriteK(".", `
resources:
  - resources.yaml

replacements:
  - source: &source
      version: v1
      kind: ConfigMap
      name: cm1
      fieldPath: data.key1
    targets:
      - select: &select
          version: v1
          kind: ConfigMap
          name: cm2
        options:
          create: true
        fieldPaths:
          - .data.keya
  - source:
      <<: *source
      fieldPath: data.key2
    targets:
      - select: *select
        options:
          create: true
        fieldPaths:
          - .data.keyb
  - source:
      <<: *source
      fieldPath: data.key2
    targets:
      - select: *select
        options:
          create: true
        fieldPaths:
          - .data.keyc
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  key1: value1
  key2: value2
  key3: value3
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
data:
  keya: value1
  keyb: value2
  keyc: value2
kind: ConfigMap
metadata:
  name: cm2
`)
}
