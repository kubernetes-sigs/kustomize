// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
)

// KustTarget is primarily tested in the krusty package with
// high level tests.

func TestLoad(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	expectedTypeMeta := types.TypeMeta{
		APIVersion: "kustomize.config.k8s.io/v1beta1",
		Kind:       "Kustomization",
	}

	testCases := map[string]struct {
		errContains string
		content     string
		k           types.Kustomization
	}{
		"empty": {
			// no content
			k: types.Kustomization{
				TypeMeta: expectedTypeMeta,
			},
		},
		"nonsenseLatin": {
			errContains: "error converting YAML to JSON",
			content: `
		Lorem ipsum dolor sit amet, consectetur
		adipiscing elit, sed do eiusmod tempor
		incididunt ut labore et dolore magna aliqua.
		Ut enim ad minim veniam, quis nostrud
		exercitation ullamco laboris nisi ut
		aliquip ex ea commodo consequat.
		`,
		},
		"simple": {
			content: `
commonLabels:
  app: nginx
`,
			k: types.Kustomization{
				TypeMeta:     expectedTypeMeta,
				CommonLabels: map[string]string{"app": "nginx"},
			},
		},
		"commented": {
			content: `
# Licensed to the Blah Blah Software Foundation
# ...
# yada yada yada.

commonLabels:
 app: nginx
`,
			k: types.Kustomization{
				TypeMeta:     expectedTypeMeta,
				CommonLabels: map[string]string{"app": "nginx"},
			},
		},
	}

	kt := makeKustTargetWithRf(
		t, th.GetFSys(), "/",
		resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			th.WriteK("/", tc.content)
			err := kt.Load()
			if tc.errContains != "" {
				require.NotNilf(t, err, "expected error containing: `%s`", tc.errContains)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.Nilf(t, err, "got error: %v", err)
				k := kt.Kustomization()
				require.Condition(t, func() bool {
					return reflect.DeepEqual(tc.k, k)
				}, "expected %v, got %v", tc.k, k)
			}
		})
	}
}

func TestMakeCustomizedResMap(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
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

	resFactory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

	resources := []*resource.Resource{
		resFactory.FromMapWithName("dply1", map[string]interface{}{
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
		resFactory.FromMapWithName("ns1", map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "ns1",
				"labels": map[string]interface{}{
					"app": "nginx",
				},
				"annotations": map[string]interface{}{
					"note": "This is a test annotation",
				},
			},
		}),
		resFactory.FromMapWithName("literalConfigMap",
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "foo-literalConfigMap-bar-g5f6t456f5",
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
		resFactory.FromMapWithName("secret",
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name":      "foo-secret-bar-82c2g5f8f6",
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

	kt := makeKustTargetWithRf(
		t, th.GetFSys(), "/whatever", resFactory)
	err := kt.Load()
	if err != nil {
		t.Fatalf("unexpected Resources error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("unexpected Resources error %v", err)
	}

	if err = expected.ErrorIfNotEqualLists(actual); err != nil {
		t.Fatalf("unexpected inequality: %v", err)
	}
}
