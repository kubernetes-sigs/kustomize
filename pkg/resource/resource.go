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

// Package resource implements representations of k8s API resources as "unstructured" objects.
package resource

import (
	"reflect"
	"strings"

	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/types"
	"sigs.k8s.io/yaml"
)

// Resource is map representation of a Kubernetes API resource object
// paired with a GenerationBehavior.
type Resource struct {
	ifc.Kunstructured
	options     *types.GenArgs
	refBy       []resid.ResId
	refVarNames []string
}

func (r *Resource) KunstructEqual(o *Resource) bool {
	return reflect.DeepEqual(r.Kunstructured, o.Kunstructured)
}

// String returns resource as JSON.
func (r *Resource) String() string {
	bs, err := r.MarshalJSON()
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return strings.TrimSpace(string(bs)) + r.options.String()
}

// DeepCopy returns a new copy of resource
func (r *Resource) DeepCopy() *Resource {
	rc := &Resource{
		Kunstructured: r.Kunstructured.Copy(),
		options:       r.options,
	}
	if len(r.refBy) > 0 {
		refby := make([]resid.ResId, len(r.refBy))
		copy(refby, r.refBy)
		rc.refBy = refby
	}
	if len(r.refVarNames) > 0 {
		refVarNames := make([]string, len(r.refVarNames))
		copy(refVarNames, r.refVarNames)
		rc.refVarNames = refVarNames
	}
	return rc
}

// AsYAML returns the resource in Yaml form.
// Easier to read than JSON.
func (r *Resource) AsYAML() ([]byte, error) {
	json, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(json)
}

// Behavior returns the behavior for the resource.
func (r *Resource) Behavior() types.GenerationBehavior {
	return r.options.Behavior()
}

// NeedHashSuffix checks if the resource need a hash suffix
func (r *Resource) NeedHashSuffix() bool {
	return r.options != nil && r.options.NeedsHashSuffix()
}

// Id returns the ResId for the resource.
func (r *Resource) Id() resid.ResId {
	namespace, _ := r.GetFieldValue("metadata.namespace")
	return resid.NewResIdWithPrefixNamespace(r.GetGvk(), r.GetName(), "", namespace)
}

// GetRefBy returns the ResIds that referred to current resource
func (r *Resource) GetRefBy() []resid.ResId {
	return r.refBy
}

// AppendRefBy appends a ResId into the refBy list
func (r *Resource) AppendRefBy(id resid.ResId) {
	r.refBy = append(r.refBy, id)
}

// GetRefVarNames returns vars that refer to current resource
func (r *Resource) GetRefVarNames() []string {
	return r.refVarNames
}

// AppendRefVarName appends a name of a var into the refVar list
func (r *Resource) AppendRefVarName(variable types.Var) {
	r.refVarNames = append(r.refVarNames, variable.Name)
}

// Merge performs merge with other resource.
func (r *Resource) Merge(other *Resource) {
	r.Replace(other)
	mergeConfigmap(r.Map(), other.Map(), r.Map())
}

// Replace performs replace with other resource.
func (r *Resource) Replace(other *Resource) {
	r.SetLabels(mergeStringMaps(other.GetLabels(), r.GetLabels()))
	r.SetAnnotations(
		mergeStringMaps(other.GetAnnotations(), r.GetAnnotations()))
	r.SetName(other.GetName())
	r.options = other.options
}

// TODO: Add BinaryData once we sync to new k8s.io/api
func mergeConfigmap(
	mergedTo map[string]interface{},
	maps ...map[string]interface{}) {
	mergedMap := map[string]interface{}{}
	for _, m := range maps {
		datamap, ok := m["data"].(map[string]interface{})
		if ok {
			for key, value := range datamap {
				mergedMap[key] = value
			}
		}
	}
	mergedTo["data"] = mergedMap
}

func mergeStringMaps(maps ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, m := range maps {
		for key, value := range m {
			result[key] = value
		}
	}
	return result
}
