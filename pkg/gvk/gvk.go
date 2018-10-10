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

package gvk

import (
	"strings"
)

// Gvk identifies a Kubernetes API type.
// https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/api-group.md
type Gvk struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// FromKind makes a Gvk with only the kind specified.
func FromKind(k string) Gvk {
	return Gvk{
		Kind: k,
	}
}

const (
	noGroup   = "noGroup"
	noVersion = "noVersion"
	noKind    = "noKind"
	separator = "_"
)

// String returns a string representation of the GVK.
func (x Gvk) String() string {
	g := x.Group
	if g == "" {
		g = noGroup
	}
	v := x.Version
	if v == "" {
		v = noVersion
	}
	k := x.Kind
	if k == "" {
		k = noKind
	}
	return strings.Join([]string{g, v, k}, separator)
}

// Equals returns true if the Gvk's have equal fields.
func (x Gvk) Equals(o Gvk) bool {
	return x.Group == o.Group && x.Version == o.Version && x.Kind == o.Kind
}

// An attempt to order things to help k8s, e.g.
// a Service should come before things that refer to it.
// Namespace should be first.
// In some cases order just specified to provide determinism.
var order = []string{
	"Namespace",
	"CustomResourceDefinition",
	"ServiceAccount",
	"Role",
	"ClusterRole",
	"RoleBinding",
	"ClusterRoleBinding",
	"ConfigMap",
	"Secret",
	"Service",
	"Deployment",
	"StatefulSet",
	"CronJob",
	"PodDisruptionBudget",
}
var typeOrders = func() map[string]int {
	m := map[string]int{}
	for i, n := range order {
		m[n] = i
	}
	return m
}()

// IsLessThan returns true if self is less than the argument.
func (x Gvk) IsLessThan(o Gvk) bool {
	indexI, foundI := typeOrders[x.Kind]
	indexJ, foundJ := typeOrders[o.Kind]
	if foundI && foundJ {
		return indexI < indexJ
	}
	if foundI && !foundJ {
		return true
	}
	if !foundI && foundJ {
		return false
	}
	return x.String() < o.String()
}

// IsSelected returns true if `selector` selects `x`; otherwise, false.
// If `selector` and `x` are the same, return true.
// If `selector` is nil, it is considered as a wildcard and always return true.
// e.g. selector <Group: "", Version: "", Kind: "Deployment"> CAN select
// <Group: "extensions", Version: "v1beta1", Kind: "Deployment">.
// selector <Group: "apps", Version: "", Kind: "Deployment"> CANNOT select
// <Group: "extensions", Version: "v1beta1", Kind: "Deployment">.
func (x Gvk) IsSelected(selector *Gvk) bool {
	if selector == nil {
		return true
	}
	if len(selector.Group) > 0 {
		if x.Group != selector.Group {
			return false
		}
	}
	if len(selector.Version) > 0 {
		if x.Version != selector.Version {
			return false
		}
	}
	if len(selector.Kind) > 0 {
		if x.Kind != selector.Kind {
			return false
		}
	}
	return true
}
