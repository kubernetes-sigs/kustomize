/*
Copyright 2017 The Kubernetes Authors.

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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Var represents a variable whose value will be sourced
// from a field in a Kubernetes object.
type Var struct {
	// Value of identifier name e.g. FOO used in container args, annotations
	// Appears in pod template as $(FOO)
	Name string `json:"name" yaml:"name"`

	// ObjRef must refer to a Kubernetes resource under the
	// purview of this kustomization. ObjRef should use the
	// raw name of the object (the name specified in its YAML,
	// before addition of a namePrefix).
	ObjRef Target `json:"objref" yaml:"objref"`

	// FieldRef refers to the field of the object referred to by
	// ObjRef whose value will be extracted for use in
	// replacing $(FOO).
	// If unspecified, this defaults to fieldpath: metadata.name
	FieldRef FieldRef `json:"fieldref,omitempty" yaml:"fieldref,omitempty"`
}

// Target represents a Kubernetes object reference
type Target struct {
	ApiVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Namespace  string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name       string `json:"name" yaml:"name"`
}

//FieldRef refers to a field of an kubernetes object
type FieldRef struct {
	FieldPath string `json:"fieldPath,omitempty" yaml:"fieldPath,omitempty"`
}

// Defaulting sets reference to field used by default.
func (v *Var) Defaulting() {
	if v.FieldRef.FieldPath == "" {
		v.FieldRef.FieldPath = "metadata.name"
	}
}

// GroupVersionKind returns a GVK
func (t Target) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(t.ApiVersion, t.Kind)
}
