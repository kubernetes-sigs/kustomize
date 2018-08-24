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

package types

import (
	"fmt"
	"reflect"
	"strings"
)

// PatchArgs represents arguments for a patch
type PatchArgs struct {
	// Relative file path within the kustomization for a json patch file.
	Path string `json:"path" yaml:"path"`

	// Type of the patch
	// The default type is Stategic Merge Patch
	// Other supported types are JSON patch and JSON Merge patch
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// Target refers to a Kubernetes object that the json patch will be
	// applied to. It must refer to a Kubernetes resource under the
	// purview of this kustomization. Target should use the
	// raw name of the object (the name specified in its YAML,
	// before addition of a namePrefix).
	Target PatchTarget `json:"target,omitempty" yaml:"target,omitempty"`
}

// PatchTarget represents the kubernetes that the patch is applied to
type PatchTarget struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Name    string `json:"name" yaml:"name"`
}

// NewPatchArgs creates a PathArgs from an interface
func NewPatchArgs(patch interface{}) (*PatchArgs, error) {
	switch patch.(type) {
	case string:
		return &PatchArgs{Path: patch.(string)}, nil
	case map[string]interface{}:
		patchArgs := &PatchArgs{}
		err := patchArgs.fillStruct(patch.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return patchArgs, nil
	default:
		return nil, fmt.Errorf("Unrecognized types from %v", patch)
	}
}

func (p *PatchArgs) fillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := setField(p, strings.Title(k), v)
		if err != nil {
			return err
		}
	}
	return nil
}

func setField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return fmt.Errorf("Provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}
