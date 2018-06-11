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

package transformers

import (
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type namespaceTransformer struct {
	namespace       string
	pathConfigs     []PathConfig
	skipPathConfigs []PathConfig
}

var namespacePathConfigs = []PathConfig{
	{
		Path:               []string{"metadata", "namespace"},
		CreateIfNotPresent: true,
	},
}

var skipNamespacePathConfigs = []PathConfig{
	{
		GroupVersionKind: &schema.GroupVersionKind{
			Kind: "Namespace",
		},
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{
			Kind: "ClusterRoleBinding",
		},
	},
	{
		GroupVersionKind: &schema.GroupVersionKind{
			Kind: "ClusterRole",
		},
	},
}

var _ Transformer = &namespaceTransformer{}

// NewNamespaceTransformer construct a namespaceTransformer.
func NewNamespaceTransformer(ns string) Transformer {
	if len(ns) == 0 {
		return NewNoOpTransformer()
	}

	return &namespaceTransformer{
		namespace:       ns,
		pathConfigs:     namespacePathConfigs,
		skipPathConfigs: skipNamespacePathConfigs,
	}
}

// Transform adds the namespace.
func (o *namespaceTransformer) Transform(m resmap.ResMap) error {
	mf := resmap.ResMap{}

	for id := range m {
		mf[id] = m[id]
		for _, path := range o.skipPathConfigs {
			if selectByGVK(id.Gvk(), path.GroupVersionKind) {
				delete(mf, id)
				break
			}
		}
	}

	for id := range mf {
		objMap := m[id].UnstructuredContent()
		for _, path := range o.pathConfigs {
			if !selectByGVK(id.Gvk(), path.GroupVersionKind) {
				continue
			}

			err := mutateField(objMap, path.Path, path.CreateIfNotPresent, func(_ interface{}) (interface{}, error) {
				return o.namespace, nil
			})
			if err != nil {
				return err
			}
		}

	}
	return nil
}
