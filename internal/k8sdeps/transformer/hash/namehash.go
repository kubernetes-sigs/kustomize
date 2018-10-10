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

package hash

import (
	"fmt"

	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
)

// nameHashTransformer contains the prefix and the path config for each field that
// the name prefix will be applied.
type nameHashTransformer struct{}

var _ transformers.Transformer = &nameHashTransformer{}

// NewNameHashTransformer construct a nameHashTransformer.
func NewNameHashTransformer() transformers.Transformer {
	return &nameHashTransformer{}
}

// Transform appends hash to configmaps and secrets.
func (o *nameHashTransformer) Transform(m resmap.ResMap) error {
	for _, res := range m {
		if res.IsGenerated() {
			err := o.appendHash(res)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *nameHashTransformer) appendHash(res *resource.Resource) error {
	h, err := NewKustHash().Hash(res.Map())
	if err != nil {
		return err
	}
	nameWithHash := fmt.Sprintf("%s-%s", res.GetName(), h)
	res.SetName(nameWithHash)
	return nil
}
