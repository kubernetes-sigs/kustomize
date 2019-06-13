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
	"sigs.k8s.io/kustomize/pkg/types"

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

func (fs *UnstructAdapter) selectSubtree(path string) (map[string]interface{}, []string, bool, error) {
	sections, err := parseFields(path)
	if len(sections) == 0 || (err != nil) {
		return nil, nil, false, err
	}

	content := fs.UnstructuredContent()
	lastSectionIdx := len(sections)

	// There are multiple sections to walk
	for sectionIdx := 0; sectionIdx < lastSectionIdx; sectionIdx++ {
		idx := sections[sectionIdx].idx
		fields := sections[sectionIdx].fields

		if idx == nil {
			// This section has no index
			return content, fields, true, nil
		}

		// This section is terminated by an indexed field.
		// Let's extract the slice first
		indexedField, found, err := unstructured.NestedFieldNoCopy(content, fields...)
		if !found || err != nil {
			return content, fields, found, err
		}
		s, ok := indexedField.([]interface{})
		if !ok {
			return content, fields, false, fmt.Errorf("%v is of the type %T, expected []interface{}", indexedField, indexedField)
		}
		if *idx >= len(s) {
			return content, fields, false, fmt.Errorf("index %d is out of bounds", *idx)
		}

		if sectionIdx == lastSectionIdx-1 {
			// This is the last section. Let's build a fake map
			// to let the rest of the field extraction to work.
			idxstring := fmt.Sprintf("[%v]", *idx)
			newContent := map[string]interface{}{idxstring: s[*idx]}
			newFields := []string{idxstring}
			return newContent, newFields, true, nil
		}

		newContent, ok := s[*idx].(map[string]interface{})
		if !ok {
			// Only map are supported here
			return content, fields, false,
				fmt.Errorf("%#v is expected to be of type map[string]interface{}", s[*idx])
		}
		content = newContent
	}

	// It seems to be an invalid path
	return nil, []string{}, false, nil
}

// GetFieldValue returns the value at the given fieldpath.
func (fs *UnstructAdapter) GetFieldValue(path string) (interface{}, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return nil, types.NoFieldError{Field: path}
	}

	s, found, err := unstructured.NestedFieldNoCopy(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return nil, types.NoFieldError{Field: path}
}

// GetString returns value at the given fieldpath.
func (fs *UnstructAdapter) GetString(path string) (string, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return "", types.NoFieldError{Field: path}
	}

	s, found, err := unstructured.NestedString(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return "", types.NoFieldError{Field: path}
}

// GetStringSlice returns value at the given fieldpath.
func (fs *UnstructAdapter) GetStringSlice(path string) ([]string, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return []string{}, types.NoFieldError{Field: path}
	}

	s, found, err := unstructured.NestedStringSlice(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return []string{}, types.NoFieldError{Field: path}
}

// GetBool returns value at the given fieldpath.
func (fs *UnstructAdapter) GetBool(path string) (bool, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return false, types.NoFieldError{Field: path}
	}

	s, found, err := unstructured.NestedBool(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return false, types.NoFieldError{Field: path}
}

// GetFloat64 returns value at the given fieldpath.
func (fs *UnstructAdapter) GetFloat64(path string) (float64, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return 0, err
	}

	s, found, err := unstructured.NestedFloat64(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return 0, types.NoFieldError{Field: path}
}

// GetInt64 returns value at the given fieldpath.
func (fs *UnstructAdapter) GetInt64(path string) (int64, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return 0, types.NoFieldError{Field: path}
	}

	s, found, err := unstructured.NestedInt64(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return 0, types.NoFieldError{Field: path}
}

// GetSlice returns value at the given fieldpath.
func (fs *UnstructAdapter) GetSlice(path string) ([]interface{}, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return nil, types.NoFieldError{Field: path}
	}

	s, found, err := unstructured.NestedSlice(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return nil, types.NoFieldError{Field: path}
}

// GetStringMap returns value at the given fieldpath.
func (fs *UnstructAdapter) GetStringMap(path string) (map[string]string, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return nil, types.NoFieldError{Field: path}
	}

	s, found, err := unstructured.NestedStringMap(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return nil, types.NoFieldError{Field: path}
}

// GetMap returns value at the given fieldpath.
func (fs *UnstructAdapter) GetMap(path string) (map[string]interface{}, error) {
	content, fields, found, err := fs.selectSubtree(path)
	if !found || err != nil {
		return nil, types.NoFieldError{Field: path}
	}

	s, found, err := unstructured.NestedMap(
		content, fields...)
	if found || err != nil {
		return s, err
	}
	return nil, types.NoFieldError{Field: path}
}
