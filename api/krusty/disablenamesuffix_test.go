// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func findSecret(m resmap.ResMap, prefix string) *resource.Resource {
	for _, r := range m.Resources() {
		if r.OrgId().Kind == "Secret" {
			if prefix == "" || strings.HasPrefix(r.GetName(), prefix) {
				return r
			}
		}
	}
	return nil
}

func TestDisableNameSuffixHash(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	const kustomizationContent = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: foo-
nameSuffix: -bar
namespace: ns1
commonLabels:
  app: nginx
commonAnnotations:
  note: This is a test annotation
resources:
  - deployment.yaml
  - namespace.yaml
generatorOptions:
  disableNameSuffixHash: false
configMapGenerator:
- name: literalConfigMap
  literals:
  - DB_USERNAME=admin
  - DB_PASSWORD=somepw
secretGenerator:
- name: secret
  literals:
    - DB_USERNAME=admin
    - DB_PASSWORD=somepw
  type: Opaque
patchesJson6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: dply1
  path: jsonpatch.json
`

	th.WriteK("/whatever", kustomizationContent)
	th.WriteF("/whatever/deployment.yaml", `
apiVersion: apps/v1
metadata:
  name: dply1
kind: Deployment
`)
	th.WriteF("/whatever/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
`)
	th.WriteF("/whatever/jsonpatch.json", `[
    {"op": "add", "path": "/spec/replica", "value": "3"}
]`)

	m := th.Run("/whatever", th.MakeDefaultOptions())

	secret := findSecret(m, "")
	if secret == nil {
		t.Errorf("Expected to find a Secret")
	}
	if secret.GetName() != "foo-secret-bar-82c2g5f8f6" {
		t.Errorf("unexpected secret resource name: %s", secret.GetName())
	}

	th.WriteK("/whatever",
		strings.Replace(kustomizationContent,
			"disableNameSuffixHash: false",
			"disableNameSuffixHash: true", -1))
	m = th.Run("/whatever", th.MakeDefaultOptions())
	secret = findSecret(m, "")
	if secret == nil {
		t.Errorf("Expected to find a Secret")
	}
	if secret.GetName() != "foo-secret-bar" { // No hash at end.
		t.Errorf("unexpected secret resource name: %s", secret.GetName())
	}
}
func TestDisableNameSuffixHashPerObject(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	const kustomizationContent = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generatorOptions:
  disableNameSuffixHash: false
secretGenerator:
- name: nohash
  options:
    disableNameSuffixHash: true
  literals:
    - DB_USERNAME=admin
    - DB_PASSWORD=somepw
  type: Opaque
- name: yeshash
  options:
    disableNameSuffixHash: false
  literals:
    - DB_USERNAME=admin
    - DB_PASSWORD=somepw
  type: Opaque
`

	th.WriteK("/whatever", kustomizationContent)

	m := th.Run("/whatever", th.MakeDefaultOptions())

	secret := findSecret(m, "nohash")
	if secret == nil {
		t.Errorf("Expected to find a Secret")
	}
	if secret.GetName() != "nohash" {
		t.Errorf("unexpected secret resource name: %s", secret.GetName())
	}

	secret = findSecret(m, "yeshash")
	if secret == nil {
		t.Errorf("Expected to find a Secret")
	}
	if secret.GetName() != "yeshash-82c2g5f8f6" {
		t.Errorf("unexpected secret resource name: %s", secret.GetName())
	}
}
