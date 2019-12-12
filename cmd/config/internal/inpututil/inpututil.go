// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package inpututil

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type MapInputsEFn func(*yaml.RNode, yaml.ResourceMeta) error

func MapInputsE(inputs []*yaml.RNode, fn MapInputsEFn) error {
	for i := range inputs {
		meta, err := inputs[i].GetMeta()
		if err != nil {
			return err
		}
		if err := fn(inputs[i], meta); err != nil {
			return err
		}
	}
	return nil
}

type MapInputsFn func(*yaml.RNode, yaml.ResourceMeta) ([]*yaml.RNode, error)

//  runs the function against each input Resource, providing the parsed metadata
func MapInputs(inputs []*yaml.RNode, fn MapInputsFn) ([]*yaml.RNode, error) {
	var outputs []*yaml.RNode
	for i := range inputs {
		meta, err := inputs[i].GetMeta()
		if err != nil {
			return nil, err
		}
		o, err := fn(inputs[i], meta)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, o...)
	}
	return outputs, nil
}
