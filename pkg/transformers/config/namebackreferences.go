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

package config

import (
	"sigs.k8s.io/kustomize/pkg/gvk"
	"strings"
)

// NameBackReferences is an association between a gvk.GVK and a list
// of FieldSpec instances that could refer to it.
//
// It is used to handle name changes, and can be thought of as a
// a contact list.  If you change your own contact info (name,
// phone number, etc.), you must tell your contacts or they won't
// know about the change.
//
// For example, ConfigMaps can be used by Pods and everything that
// contains a Pod; Deployment, Job, StatefulSet, etc.  To change
// the name of a ConfigMap instance from 'alice' to 'bob', one
// must visit all objects that could refer to the ConfigMap, see if
// they mention 'alice', and if so, change the reference to 'bob'.
//
// The NameBackReferences instance to aid in this could look like
//   {
//     kind: ConfigMap
//     version: v1
//     FieldSpecs:
//     - kind: Pod
//       version: v1
//       path: spec/volumes/configMap/name
//     - kind: Deployment
//       path: spec/template/spec/volumes/configMap/name
//     - kind: Job
//       path: spec/template/spec/volumes/configMap/name
//       (etc.)
//   }
type NameBackReferences struct {
	gvk.Gvk    `json:",inline,omitempty" yaml:",inline,omitempty"`
	FieldSpecs []FieldSpec `json:"FieldSpecs,omitempty" yaml:"FieldSpecs,omitempty"`
}

func (n NameBackReferences) String() string {
	var r []string
	for _, f := range n.FieldSpecs {
		r = append(r, f.String())
	}
	return n.Gvk.String() + ":  (\n" +
		strings.Join(r, "\n") + "\n)"
}

func mergeNameBackReferences(
	a, b []NameBackReferences) []NameBackReferences {
	for _, r := range b {
		a = merge(a, r)
	}
	return a
}

func merge(
	backRefsSlice []NameBackReferences,
	other NameBackReferences) []NameBackReferences {
	var result []NameBackReferences
	found := false
	for _, c := range backRefsSlice {
		if c.Equals(other.Gvk) {
			c.FieldSpecs = append(c.FieldSpecs, other.FieldSpecs...)
			found = true
		}
		result = append(result, c)
	}

	if !found {
		result = append(result, other)
	}
	return result
}
