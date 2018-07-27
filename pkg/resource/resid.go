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
	fields := []string{n.gvk.Group, n.gvk.Version, n.gvk.Kind, n.prefix, n.name}
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

// CopyWithNewPrefix make a new copy from current ResId and append a new prefix
func (n ResId) CopyWithNewPrefix(p string) ResId {
	return ResId{gvk: n.gvk, name: n.name, prefix: p + n.prefix}
}
