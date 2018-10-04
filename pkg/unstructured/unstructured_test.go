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

package unstructured

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
)

func makeUnstructured() *Unstructured {
	obj := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name": "myapp",
		},
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{
					"name":  "pod",
					"image": "busybox",
				},
			},
		},
	}
	return &Unstructured{Object: obj}
}

func TestDeepCopy(t *testing.T) {
	u := makeUnstructured()
	copy := u.DeepCopyObject()
	if !reflect.DeepEqual(u, copy) {
		t.Fatalf("The new copy %v is not the same as the original %v", copy, u)
	}
}

func TestGroupVersionKind(t *testing.T) {
	u := makeUnstructured()
	expected := gvk.Gvk{
		Group:   "apps",
		Version: "v2",
		Kind:    "Application",
	}
	u.SetGvk(expected)
	if !reflect.DeepEqual(u.Gvk(), expected) {
		t.Fatalf("expected %v, but got %v", expected, u.Gvk())
	}
}

func TestName(t *testing.T) {
	u := makeUnstructured()
	name := "new-name"
	u.SetName(name)
	if !reflect.DeepEqual(u.GetName(), name) {
		t.Fatalf("expected %v, but got %v", name, u.GetName())
	}
}

func TestAnnotation(t *testing.T) {
	u := makeUnstructured()
	annotations := map[string]string{
		"foo": "bar",
	}
	u.SetAnnotations(annotations)
	if !reflect.DeepEqual(u.GetAnnotations(), annotations) {
		t.Fatalf("expected %v, but got %v", annotations, u.GetAnnotations())
	}
}

func TestLabel(t *testing.T) {
	u := makeUnstructured()
	labels := map[string]string{
		"bar": "foo",
	}
	u.SetLabels(labels)
	if !reflect.DeepEqual(u.GetLabels(), labels) {
		t.Fatalf("expected %v, but got %v", labels, u.GetLabels())
	}
}
