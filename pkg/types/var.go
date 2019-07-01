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
	"fmt"
	"reflect"
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
)

const defaultFieldPath = "metadata.name"

// Var represents a variable whose value will be sourced
// from a field in a Kubernetes object.
type Var struct {
	// Value of identifier name e.g. FOO used in container args, annotations
	// Appears in pod template as $(FOO)
	Name string `json:"name" yaml:"name"`

	// ObjRef must refer to a Kubernetes resource under the
	// purview of this kustomization. ObjRef should use the
	// raw name of the object (the name specified in its YAML,
	// before addition of a namePrefix and a nameSuffix).
	ObjRef Target `json:"objref" yaml:"objref"`

	// FieldRef refers to the field of the object referred to by
	// ObjRef whose value will be extracted for use in
	// replacing $(FOO).
	// If unspecified, this defaults to fieldPath: $defaultFieldPath
	FieldRef FieldSelector `json:"fieldref,omitempty" yaml:"fieldref,omitempty"`
}

// Target refers to a kubernetes object by Group, Version, Kind and Name
// gvk.Gvk contains Group, Version and Kind
// APIVersion is added to keep the backward compatibility of using ObjectReference
// for Var.ObjRef
type Target struct {
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	gvk.Gvk    `json:",inline,omitempty" yaml:",inline,omitempty"`
	Name       string `json:"name" yaml:"name"`
}

// FieldSelector contains the fieldPath to an object field.
// This struct is added to keep the backward compatibility of using ObjectFieldSelector
// for Var.FieldRef
type FieldSelector struct {
	FieldPath string `json:"fieldPath,omitempty" yaml:"fieldPath,omitempty"`
}

// defaulting sets reference to field used by default.
func (v *Var) defaulting() {
	if v.FieldRef.FieldPath == "" {
		v.FieldRef.FieldPath = defaultFieldPath
	}
}

// VarSet is a set of Vars where no var.Name is repeated.
type VarSet struct {
	set map[string]Var
}

// NewVarSet returns an initialized VarSet
func NewVarSet() VarSet {
	return VarSet{set: map[string]Var{}}
}

// AsSlice returns the vars as a slice.
func (vs *VarSet) AsSlice() []Var {
	s := make([]Var, len(vs.set))
	i := 0
	for _, v := range vs.set {
		s[i] = v
		i++
	}
	sort.Sort(ByName(s))
	return s
}

// Copy returns a copy of the var set.
func (vs *VarSet) Copy() VarSet {
	newSet := make(map[string]Var, len(vs.set))
	for k, v := range vs.set {
		newSet[k] = v
	}
	return VarSet{set: newSet}
}

// MergeSet absorbs other vars with error on name collision.
func (vs *VarSet) MergeSet(incoming VarSet) error {
	for _, incomingVar := range incoming.set {
		if err := vs.Merge(incomingVar); err != nil {
			return err
		}
	}
	return nil
}

// MergeSlice absorbs a Var slice with error on name collision.
// Empty fields in incoming vars are defaulted.
func (vs *VarSet) MergeSlice(incoming []Var) error {
	for _, v := range incoming {
		if err := vs.Merge(v); err != nil {
			return err
		}
	}
	return nil
}

// Merge absorbs another Var with error on name collision.
// Empty fields in incoming Var is defaulted.
func (vs *VarSet) Merge(v Var) error {
	v.defaulting()
	if vs.Contains(v) {
		// Only return an error if a variable with the same
		// name already exists
		if !reflect.DeepEqual(v, vs.set[v.Name]) {
			return fmt.Errorf(
				"var '%s' already encountered", v.Name)
		}
		return nil
	}
	vs.set[v.Name] = v
	return nil
}

// Contains is true if the set has the other var.
func (vs *VarSet) Contains(other Var) bool {
	return vs.Get(other.Name) != nil
}

// Get returns the var with the given name, else nil.
func (vs *VarSet) Get(name string) *Var {
	if v, found := vs.set[name]; found {
		return &v
	}
	return nil
}

// GVK returns the Gvk object in Target
func (t *Target) GVK() gvk.Gvk {
	if t.APIVersion == "" {
		return t.Gvk
	}
	versions := strings.Split(t.APIVersion, "/")
	if len(versions) == 2 {
		t.Group = versions[0]
		t.Version = versions[1]
	}
	if len(versions) == 1 {
		t.Version = versions[0]
	}
	return t.Gvk
}

// ByName is a sort interface which sorts Vars by name alphabetically
type ByName []Var

func (v ByName) Len() int           { return len(v) }
func (v ByName) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v ByName) Less(i, j int) bool { return v[i].Name < v[j].Name }
