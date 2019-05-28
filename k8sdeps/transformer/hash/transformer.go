// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"fmt"

	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers"
)

type transformer struct{}

var _ transformers.Transformer = &transformer{}

// NewTransformer make a hash transformer.
func NewTransformer() transformers.Transformer {
	return &transformer{}
}

// Transform appends hash to generated resources.
func (tf *transformer) Transform(m resmap.ResMap) error {
	for _, res := range m {
		if res.NeedHashSuffix() {
			h, err := NewKustHash().Hash(res.Map())
			if err != nil {
				return err
			}
			res.SetName(fmt.Sprintf("%s-%s", res.GetName(), h))
		}
	}
	return nil
}
