/*
Copyright 2019 The Kubernetes Authors.

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

package resid

import (
	"strings"

	"sigs.k8s.io/kustomize/pkg/gvk"
)

// ItemId  contains the group, version, kind, namespace
// and name of a resource
type ItemId struct {
	// Gvk of the resource.
	gvk.Gvk `json:",inline,omitempty" yaml:",inline,omitempty"`

	// Name of the resource before transformation.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Namespace the resource belongs to.
	// An untransformed resource has no namespace.
	// A fully transformed resource has the namespace
	// from the top most overlay.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// String of ItemId based on GVK, name and namespace
func (i ItemId) String() string {
	ns := i.Namespace
	if ns == "" {
		ns = noNamespace
	}
	nm := i.Name
	if nm == "" {
		nm = noName
	}

	return strings.Join(
		[]string{i.Gvk.String(), ns, nm}, separator)
}

func (i ItemId) Equals(b ItemId) bool {
	return i.String() == b.String()
}

func NewItemId(g gvk.Gvk, ns, nm string) ItemId {
	return ItemId{
		Gvk:       g,
		Namespace: ns,
		Name:      nm,
	}
}

func FromString(s string) ItemId {
	values := strings.Split(s, separator)
	g := gvk.FromString(values[0])

	ns := values[1]
	if ns == noNamespace {
		ns = ""
	}
	nm := values[2]
	if nm == noName {
		nm = ""
	}
	return ItemId{
		Gvk:       g,
		Namespace: ns,
		Name:      nm,
	}
}
