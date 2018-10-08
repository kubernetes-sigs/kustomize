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

// Package unstructured defines the Unstructured type
package unstructured

import (
	"encoding/json"
	"strings"

	"sigs.k8s.io/kustomize/pkg/gvk"
)

// Unstructured object
type Unstructured struct {
	// Object is a JSON compatible map with string, float, int, bool, []interface{}, or
	// map[string]interface{}
	// children.
	Object map[string]interface{}
}

// DeepCopyObject returns a new copy of an Unstructured
func (u *Unstructured) DeepCopyObject() *Unstructured {
	data, _ := json.Marshal(u.Object)
	result := &Unstructured{}
	json.Unmarshal(data, result)
	return result
}

// UnmarshalJSON unmarshal JSON format bytes to an Unstructured object
func (u *Unstructured) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &u.Object)
}

// MarshalJSON marshal an unstructured into byte slice
func (u *Unstructured) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Object)
}

// Gvk returns the Group, Version and Kind of an unstructured
func (u *Unstructured) Gvk() gvk.Gvk {
	var group, version string
	apiVersion, _ := u.GetFieldValue("apiVersion")
	kind, _ := u.GetFieldValue("kind")
	versions := strings.Split(apiVersion, "/")
	if len(versions) == 1 {
		version = versions[0]
	}
	if len(versions) == 2 {
		group = versions[0]
		version = versions[1]
	}
	return gvk.Gvk{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
}

// SetGvk sets the Group, Version and Kind of an unstructured
func (u *Unstructured) SetGvk(kind gvk.Gvk) {
	apiVersion := kind.Version
	if kind.Group != "" {
		apiVersion = kind.Group + "/" + kind.Version
	}
	u.SetFieldValue("apiVersion", apiVersion)
	u.SetFieldValue("kind", kind.Kind)
}

// GetName returns metadata.name
func (u *Unstructured) GetName() string {
	name, err := u.GetFieldValue("metadata.name")
	if err != nil {
		return ""
	}
	return name
}

// SetName sets a new value for metadata.name
func (u *Unstructured) SetName(n string) {
	u.SetFieldValue("metadata.name", n)
}

// GetLabels returns metadata.label
func (u *Unstructured) GetLabels() map[string]string {
	labels, _ := u.GetMapFieldValue("metadata.labels")
	l := convertToStringMap(labels)
	return l
}

// SetLabels sets a set of labels for metadata.label
func (u *Unstructured) SetLabels(m map[string]string) {
	u.SetMapFieldValue("metadata.labels", convertToInterfaceMap(m))
}

// GetAnnotations returns metadata.annotations
func (u *Unstructured) GetAnnotations() map[string]string {
	annotations, _ := u.GetMapFieldValue("metadata.annotations")
	a := convertToStringMap(annotations)
	return a
}

// SetAnnotations sets a set of annotations for metadata.annotations
func (u *Unstructured) SetAnnotations(m map[string]string) {
	u.SetMapFieldValue("metadata.annotations", convertToInterfaceMap(m))
}

// GetFieldValue returns value at the given fieldpath.
func (u *Unstructured) GetFieldValue(fieldPath string) (string, error) {
	return getFieldValue(u.Object, strings.Split(fieldPath, "."))
}

// SetFieldValue set value at the given fieldpath.
func (u *Unstructured) SetFieldValue(fieldPath, value string) error {
	return mutateField(u.Object, strings.Split(fieldPath, "."), true, value)
}

// GetMapFieldValue returns value at the given fieldpath.
func (u *Unstructured) GetMapFieldValue(fieldPath string) (map[string]interface{}, error) {
	return getMapFieldValue(u.Object, strings.Split(fieldPath, "."))
}

// SetMapFieldValue sets a map to the given fieldpath
func (u *Unstructured) SetMapFieldValue(fieldPath string, m map[string]interface{}) error {
	return setMapFieldValue(u.Object, strings.Split(fieldPath, "."), m)
}
