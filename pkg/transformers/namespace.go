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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/pkg/resmap"
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
	{
		GroupVersionKind: &schema.GroupVersionKind{
			Kind: "CustomResourceDefinition",
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
		found := false
		for _, path := range o.skipPathConfigs {
			if selectByGVK(id.Gvk(), path.GroupVersionKind) {
				found = true
				break
			}
		}
		if !found {
			mf[id] = m[id]
			delete(m, id)
		}
	}

	for id := range mf {
		objMap := mf[id].UnstructuredContent()
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
			newid := id.CopyWithNewNamespace(o.namespace)
			m[newid] = mf[id]
		}

	}
	o.updateClusterRoleBinding(m)
	return nil
}

func (o *namespaceTransformer) updateClusterRoleBinding(m resmap.ResMap) {
	saMap := map[string]bool{}
	saGVK := schema.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"}
	for id := range m {
		if id.Gvk().String() == saGVK.String() {
			saMap[id.Name()] = true
		}
	}

	for id := range m {
		if id.Gvk().Kind != "ClusterRoleBinding" && id.Gvk().Kind != "RoleBinding" {
			continue
		}
		objMap := m[id].UnstructuredContent()
		subjects := objMap["subjects"].([]interface{})
		for i := range subjects {
			subject := subjects[i].(map[string]interface{})
			kind, foundk := subject["kind"]
			name, foundn := subject["name"]
			if !foundk || !foundn || kind.(string) != "ServiceAccount" {
				continue
			}
			// a ServiceAccount named “default” exists in every active namespace
			if name.(string) == "default" || saMap[name.(string)] {
				subject := subjects[i].(map[string]interface{})
				mutateField(subject, []string{"namespace"}, true, func(_ interface{}) (interface{}, error) {
					return o.namespace, nil
				})
				subjects[i] = subject
			}
		}
		objMap["subjects"] = subjects
	}
}
