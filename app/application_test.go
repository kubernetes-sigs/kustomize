/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/kustomize/resource"
	"k8s.io/kubectl/pkg/kustomize/types"
	"k8s.io/kubectl/pkg/loader"
	"k8s.io/kubectl/pkg/loader/loadertest"
)

func setupTest(t *testing.T) loader.Loader {
	kustomizationContent := []byte(`kustomizationName: nginx-app
namePrefix: foo-
objectLabels:
  app: nginx
objectAnnotations:
  note: This is a test annotation
resources:
  - deployment.yaml
configMapGenerator:
- name: literalConfigMap
  literals:
  - DB_USERNAME=admin
  - DB_PASSWORD=somepw
secretGenerator:
- name: secret
  commands:
    DB_USERNAME: "printf admin"
    DB_PASSWORD: "printf somepw"
  type: Opaque
`)
	deploymentContent := []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: dply1
`)

	loader := loadertest.NewFakeLoader("/testpath")
	err := loader.AddFile("/testpath/kustomize.yaml", kustomizationContent)
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	err = loader.AddFile("/testpath/deployment.yaml", deploymentContent)
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	return loader
}

func TestResources(t *testing.T) {
	expected := resource.ResourceCollection{
		types.GroupVersionKindName{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "dply1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "foo-dply1",
						"labels": map[string]interface{}{
							"app": "nginx",
						},
						"annotations": map[string]interface{}{
							"note": "This is a test annotation",
						},
					},
					"spec": map[string]interface{}{
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
				},
			},
		},
		types.GroupVersionKindName{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Name: "literalConfigMap",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "foo-literalConfigMap-mc92bgcbh5",
						"labels": map[string]interface{}{
							"app": "nginx",
						},
						"annotations": map[string]interface{}{
							"note": "This is a test annotation",
						},
						"creationTimestamp": nil,
					},
					"data": map[string]interface{}{
						"DB_USERNAME": "admin",
						"DB_PASSWORD": "somepw",
					},
				},
			},
		},
		types.GroupVersionKindName{
			GVK:  schema.GroupVersionKind{Version: "v1", Kind: "Secret"},
			Name: "secret",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]interface{}{
						"name": "foo-secret-877fcfhgt5",
						"labels": map[string]interface{}{
							"app": "nginx",
						},
						"annotations": map[string]interface{}{
							"note": "This is a test annotation",
						},
						"creationTimestamp": nil,
					},
					"type": string(corev1.SecretTypeOpaque),
					"data": map[string]interface{}{
						"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
						"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
					},
				},
			},
		},
	}
	l := setupTest(t)
	app, err := New(l)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	actual, err := app.Resources()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		err = compareMap(actual, expected)
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRawResources(t *testing.T) {
	expected := resource.ResourceCollection{
		types.GroupVersionKindName{
			GVK:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Name: "dply1",
		}: &resource.Resource{
			Data: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "dply1",
					},
				},
			},
		},
	}
	l := setupTest(t)
	app, err := New(l)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	actual, err := app.RawResources()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if err := compareMap(actual, expected); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func compareMap(m1, m2 resource.ResourceCollection) error {
	if len(m1) != len(m2) {
		keySet1 := []types.GroupVersionKindName{}
		keySet2 := []types.GroupVersionKindName{}
		for GVKn := range m1 {
			keySet1 = append(keySet1, GVKn)
		}
		for GVKn := range m1 {
			keySet2 = append(keySet2, GVKn)
		}
		return fmt.Errorf("maps has different number of entries: %#v doesn't equals %#v", keySet1, keySet2)
	}
	for GVKn, obj1 := range m1 {
		obj2, found := m2[GVKn]
		if !found {
			return fmt.Errorf("%#v doesn't exist in %#v", GVKn, m2)
		}
		if !reflect.DeepEqual(obj1, obj2) {
			return fmt.Errorf("%#v doesn't match %#v", obj1, obj2)
		}
	}
	return nil
}
