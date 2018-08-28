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

import (
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewPatchJson6902 load a slice of PatchJson6902
func NewPatchJson6902(l loader.Loader, slice []PatchJson6902) (map[resource.ResId][]byte, error) {
	patches := make(map[resource.ResId][]byte)
	for _, p := range slice {
		id := resource.NewResIdWithPrefixNamespace(
			schema.GroupVersionKind{
				Group:   p.Target.Group,
				Version: p.Target.Version,
				Kind:    p.Target.Kind,
			},
			p.Target.Name,
			"",
			p.Target.Namespace,
		)
		content, err := l.Load(p.Path)
		if err != nil {
			return nil, err
		}
		if val, ok := patches[id]; ok {
			content = append(val, content...)
		}
		patches[id] = content
	}
	return patches, nil
}
