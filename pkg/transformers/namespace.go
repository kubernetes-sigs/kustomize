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
)

type namespaceTransformer struct {
	namespace string
}

var _ Transformer = &namespaceTransformer{}

// NewNamespaceTransformer construct a namespaceTransformer.
func NewNamespaceTransformer(ns string) Transformer {
	if len(ns) == 0 {
		return NewNoOpTransformer()
	}
	return &namespaceTransformer{namespace: ns}
}

// Transform adds the namespace.
func (o *namespaceTransformer) Transform(m resmap.ResMap) error {
	for _, obj := range m {
		obj.Unstruct().SetNamespace(o.namespace)
	}
	return nil
}
