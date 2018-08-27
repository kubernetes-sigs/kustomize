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

package patch

// PatchJson6902 represents a json patch for an object
// with format documented https://tools.ietf.org/html/rfc6902.
type PatchJson6902 struct {
	// Relative file path within the kustomization for a json patch file.
	Path string `json:"path" yaml:"path"`

	// Target refers to a Kubernetes object that the json patch will be
	// applied to. It must refer to a Kubernetes resource under the
	// purview of this kustomization. Target should use the
	// raw name of the object (the name specified in its YAML,
	// before addition of a namePrefix).
	Target Target `json:"target" yaml:"target"`
}

// Target represents the kubernetes object that the patch is applied to
type Target struct {
	Group     string `json:"group,omitempty" yaml:"group,omitempty"`
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind      string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name      string `json:"name" yaml:"name"`
}
