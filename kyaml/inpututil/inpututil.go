// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package inpututil

import (
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type MapInputsEFn func(*yaml.RNode, yaml.ResourceMeta) error

// MapInputsE runs the function against each input Resource, providing the parsed metadata
func MapInputsE(inputs []*yaml.RNode, fn MapInputsEFn) error {
	for i := range inputs {
		meta, err := inputs[i].GetMeta()
		if err != nil {
			return errors.Wrap(err)
		}
		if err := fn(inputs[i], meta); err != nil {
			return WrapErrorWithFile(err, meta)
		}
	}
	return nil
}

type MapInputsFn func(*yaml.RNode, yaml.ResourceMeta) ([]*yaml.RNode, error)

// MapInputs runs the function against each input Resource, providing the parsed metadata
// and returning the aggregated result
func MapInputs(inputs []*yaml.RNode, fn MapInputsFn) ([]*yaml.RNode, error) {
	var outputs []*yaml.RNode
	for i := range inputs {
		meta, err := inputs[i].GetMeta()
		if err != nil {
			return nil, errors.Wrap(err)
		}
		o, err := fn(inputs[i], meta)
		if err != nil {
			return nil, WrapErrorWithFile(err, meta)
		}
		outputs = append(outputs, o...)
	}
	return outputs, nil
}

// WrapErrorWithFile returns the original error wrapped with information about the file
// that the Resource was parsed from.
func WrapErrorWithFile(err error, meta yaml.ResourceMeta) error {
	if err == nil {
		return err
	}
	return errors.WrapPrefixf(err, "%s [%s]",
		meta.Annotations[kioutil.PathAnnotation],
		meta.Annotations[kioutil.IndexAnnotation])
}
