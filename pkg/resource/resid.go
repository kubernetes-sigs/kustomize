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

package resource

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResId conflates GroupVersionKind with a textual name to uniquely identify a kubernetes resource (object).
type ResId struct {
	// GroupVersionKind of the resource.
	gvk schema.GroupVersionKind
	// original name of the resource before transformation.
	name string
	// namePrefix of the resource
	// an untransformed resource has no prefix, fully transformed resource has an arbitrary number of prefixes
	// concatenated together.
	prefix string
	// namespace the resource belongs to
	// an untransformed resource has no namespace, fully transformed resource has the namespace from
	// the top most overlay
	namespace string
}

// NewResIdWithPrefixNamespace creates new resource identifier with a prefix and a namespace
func NewResIdWithPrefixNamespace(g schema.GroupVersionKind, n, p, ns string) ResId {
	return ResId{gvk: g, name: n, prefix: p, namespace: ns}
}

// NewResIdWithPrefix creates new resource identifier with a prefix
func NewResIdWithPrefix(g schema.GroupVersionKind, n, p string) ResId {
	return ResId{gvk: g, name: n, prefix: p}
}

// NewResId creates new resource identifier
func NewResId(g schema.GroupVersionKind, n string) ResId {
	return NewResIdWithPrefix(g, n, "")
}

// String of ResId based on GVK, name and prefix
func (n ResId) String() string {
	fields := []string{n.gvk.Group, n.gvk.Version, n.gvk.Kind, n.namespace, n.prefix, n.name}
	return strings.Join(fields, "_") + ".yaml"
}

// GvknString of ResId based on GVK and name
func (n ResId) GvknString() string {
	if n.gvk.Group == "" {
		return strings.Join([]string{n.gvk.Version, n.gvk.Kind, n.name}, "_") + ".yaml"
	}
	return strings.Join([]string{n.gvk.Group, n.gvk.Version, n.gvk.Kind, n.name}, "_") + ".yaml"

}

// GvknEquals return if two ResId have the same Group/Version/Kind and name
// The comparison excludes prefix
func (n ResId) GvknEquals(id ResId) bool {
	return n.gvk.Group == id.gvk.Group && n.gvk.Version == id.gvk.Version &&
		n.gvk.Kind == id.gvk.Kind && n.name == id.name
}

// Gvk returns Group/Version/Kind of the resource.
func (n ResId) Gvk() schema.GroupVersionKind {
	return n.gvk
}

// Name returns resource name.
func (n ResId) Name() string {
	return n.name
}

// Prefix returns name prefix.
func (n ResId) Prefix() string {
	return n.prefix
}

// Namespace returns resource namespace.
func (n ResId) Namespace() string {
	return n.namespace
}

// CopyWithNewPrefix make a new copy from current ResId and append a new prefix
func (n ResId) CopyWithNewPrefix(p string) ResId {
	return ResId{gvk: n.gvk, name: n.name, prefix: n.concatPrefix(p), namespace: n.namespace}
}

// CopyWithNewNamespace make a new copy from current ResId and set a new namespace
func (n ResId) CopyWithNewNamespace(ns string) ResId {
	return ResId{gvk: n.gvk, name: n.name, prefix: n.prefix, namespace: ns}
}

// HasSameLeftmostPrefix check if two ResIds have the same
// left most prefix.
func (n ResId) HasSameLeftmostPrefix(id ResId) bool {
	prefixes1 := n.prefixList()
	prefixes2 := id.prefixList()
	return prefixes1[0] == prefixes2[0]
}

func (n ResId) concatPrefix(p string) string {
	if p == "" {
		return n.prefix
	}
	if n.prefix == "" {
		return p
	}
	return p + ":" + n.prefix
}

func (n ResId) prefixList() []string {
	return strings.Split(n.prefix, ":")
}
