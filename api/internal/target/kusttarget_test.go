// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"encoding/base64"
	"testing"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// KustTarget is primarily tested in the krusty package with
// high level tests.

func TestMakeCustomizedResMap(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/whatever")
	th.WriteK("/whatever", `
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
`)
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

	resources := []*resource.Resource{
		th.RF().FromMapWithName("dply1", map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "foo-dply1-bar",
				"namespace": "ns1",
				"labels": map[string]interface{}{
					"app": "nginx",
				},
				"annotations": map[string]interface{}{
					"note": "This is a test annotation",
				},
			},
			"spec": map[string]interface{}{
				"replica": "3",
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "nginx",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"note": "This is a test annotation",
						},
						"labels": map[string]interface{}{
							"app": "nginx",
						},
					},
				},
			},
		}),
		th.RF().FromMapWithName("ns1", map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "foo-ns1-bar",
				"labels": map[string]interface{}{
					"app": "nginx",
				},
				"annotations": map[string]interface{}{
					"note": "This is a test annotation",
				},
			},
		}),
		th.RF().FromMapWithName("literalConfigMap",
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "foo-literalConfigMap-bar-8d2dkb8k24",
					"namespace": "ns1",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
				},
				"data": map[string]interface{}{
					"DB_USERNAME": "admin",
					"DB_PASSWORD": "somepw",
				},
			}),
		th.RF().FromMapWithName("secret",
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name":      "foo-secret-bar-9btc7bt4kb",
					"namespace": "ns1",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
				},
				"type": ifc.SecretTypeOpaque,
				"data": map[string]interface{}{
					"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
					"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
				},
			}),
	}

	expected := resmap.New()
	for _, r := range resources {
		if err := expected.Append(r); err != nil {
			t.Fatalf("unexpected error %v", err)
		}
	}

	actual, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("unexpected Resources error %v", err)
	}

	if err = expected.ErrorIfNotEqualLists(actual); err != nil {
		t.Fatalf("unexpected inequality: %v", err)
	}
}
