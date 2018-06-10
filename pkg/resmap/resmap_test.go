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

package resmap

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/internal/loadertest"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var deploy = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
var statefulset = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}

func TestEncodeAsYaml(t *testing.T) {
	encoded := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
`)
	input := ResMap{
		resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
		resource.NewResId(cmap, "cm2"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm2",
				},
			}),
	}
	out, err := input.EncodeAsYaml()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out, encoded) {
		t.Fatalf("%s doesn't match expected %s", out, encoded)
	}
}

func TestNewMapFromFiles(t *testing.T) {

	resourceStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: dply1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dply2
`

	l := loadertest.NewFakeLoader("/home/seans/project")
	if ferr := l.AddFile("/home/seans/project/deployment.yaml", []byte(resourceStr)); ferr != nil {
		t.Fatalf("Error adding fake file: %v\n", ferr)
	}
	expected := ResMap{resource.NewResId(deploy, "dply1"): resource.NewResourceFromMap(
		map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "dply1",
			},
		}),
		resource.NewResId(deploy, "dply2"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "dply2",
				},
			}),
	}

	m, _ := NewResMapFromFiles(l, []string{"/home/seans/project/deployment.yaml"})
	if len(m) != 2 {
		t.Fatalf("%#v should contain 2 appResource, but got %d", m, len(m))
	}

	if err := expected.ErrorIfNotEqual(m); err != nil {
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestNewMapFromBytes(t *testing.T) {
	encoded := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
`)
	expected := ResMap{
		resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
		resource.NewResId(cmap, "cm2"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm2",
				},
			}),
	}
	m, err := newResMapFromBytes(encoded)
	fmt.Printf("%v\n", m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		t.Fatalf("%#v doesn't match expected %#v", m, expected)
	}
}

func TestMerge(t *testing.T) {
	input1 := ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "foo-deploy1",
				},
			}),
	}
	input2 := ResMap{
		resource.NewResId(statefulset, "stateful1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "StatefulSet",
				"metadata": map[string]interface{}{
					"name": "bar-stateful",
				},
			}),
	}
	input := []ResMap{input1, input2}
	expected := ResMap{
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "foo-deploy1",
				},
			}),
		resource.NewResId(statefulset, "stateful1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "StatefulSet",
				"metadata": map[string]interface{}{
					"name": "bar-stateful",
				},
			}),
	}
	merged, err := Merge(input...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(merged, expected) {
		t.Fatalf("%#v doesn't equal expected %#v", merged, expected)
	}
}
