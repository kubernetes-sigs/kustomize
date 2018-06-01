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
}

func NewResId(g schema.GroupVersionKind, n string) ResId {
	return ResId{gvk: g, name: n}
}

func (n ResId) String() string {
	if n.gvk.Group == "" {
		return strings.Join([]string{n.gvk.Version, n.gvk.Kind, n.name}, "_") + ".yaml"
	}
	return strings.Join([]string{n.gvk.Group, n.gvk.Version, n.gvk.Kind, n.name}, "_") + ".yaml"
}

func (n ResId) Gvk() schema.GroupVersionKind {
	return n.gvk
}

func (n ResId) Name() string {
	return n.name
}
