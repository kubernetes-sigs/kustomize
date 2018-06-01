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

import "github.com/kubernetes-sigs/kustomize/pkg/resmap"

// multiTransformer contains a list of transformers.
type multiTransformer struct {
	transformers []Transformer
}

var _ Transformer = &multiTransformer{}

// NewMultiTransformer constructs a multiTransformer.
func NewMultiTransformer(t []Transformer) Transformer {
	r := &multiTransformer{
		transformers: make([]Transformer, len(t))}
	copy(r.transformers, t)
	return r
}

// Transform prepends the name prefix.
func (o *multiTransformer) Transform(m resmap.ResMap) error {
	for _, t := range o.transformers {
		err := t.Transform(m)
		if err != nil {
			return err
		}
	}
	return nil
}
