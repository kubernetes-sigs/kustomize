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

// Package kunstruct provides unstructured from api machinery and factory for creating unstructured
package kunstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

var _ ifc.Kunstructured = &UnstructAdapter{}

// UnstructAdapter wraps unstructured.Unstructured from
// https://github.com/kubernetes/apimachinery/blob/master/
//     pkg/apis/meta/v1/unstructured/unstructured.go
// to isolate dependence on apimachinery.
type UnstructAdapter struct {
	unstructured.Unstructured
}

// NewKunstructuredFromObject returns a new instance of Kunstructured.
func NewKunstructuredFromObject(obj runtime.Object) (ifc.Kunstructured, error) {
	// Convert obj to a byte stream, then convert that to JSON (Unstructured).
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return &UnstructAdapter{}, err
	}
	var u unstructured.Unstructured
	err = u.UnmarshalJSON(marshaled)
	// creationTimestamp always 'null', remove it
	u.SetCreationTimestamp(metav1.Time{})
	return &UnstructAdapter{Unstructured: u}, err
}

// GetGvk returns the Gvk name of the object.
func (fs *UnstructAdapter) GetGvk() gvk.Gvk {
	x := fs.GroupVersionKind()
	return gvk.Gvk{
		Group:   x.Group,
		Version: x.Version,
		Kind:    x.Kind,
	}
}

// Copy provides a copy behind an interface.
func (fs *UnstructAdapter) Copy() ifc.Kunstructured {
	return &UnstructAdapter{*fs.DeepCopy()}
}

// Map returns the unstructured content map.
func (fs *UnstructAdapter) Map() map[string]interface{} {
	return fs.Object
}

// SetMap overrides the unstructured content map.
func (fs *UnstructAdapter) SetMap(m map[string]interface{}) {
	fs.Object = m
}

// GetFieldValue returns value at the given fieldpath.
func (fs *UnstructAdapter) GetFieldValue(path string) (string, error) {
	s, err := getFieldValue(fs.UnstructuredContent(), strings.Split(path, "."))
	if err != nil {
		return "", errors.Wrapf(err, "at path '%s'", path)
	}
	return s, nil
}

func getFieldValue(m map[string]interface{}, path []string) (string, error) {
	if len(path) == 0 {
		return "", fmt.Errorf("%v not found", path)
	}
	v, ok := m[path[0]]
	if !ok {
		return "", fmt.Errorf("no field named '%s'", path[0])
	}
	if len(path) == 1 {
		if s, ok := v.(string); ok {
			return s, nil
		}
		return "", fmt.Errorf("value at '%v' not a string", path[0])
	}
	if deeper, ok := v.(map[string]interface{}); ok {
		return getFieldValue(deeper, path[1:])
	}
	return "", fmt.Errorf("Expected map at %v, but got %s=%v",
		path[0], reflect.TypeOf(v), v)
}
