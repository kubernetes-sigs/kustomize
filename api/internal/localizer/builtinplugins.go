// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// localizeBuiltinGenerators localizes built-in generators with file paths.
// Note that this excludes helm, which needs a repo.
type localizeBuiltinGenerators struct {
}

var _ kio.Filter = &localizeBuiltinGenerators{}

// Filter localizes the built-in generators with file paths. Filter returns an error if
// generators contains a resource that is not a built-in generator, cannot contain a file path,
// needs more than a file path like helm, or is not localizable.
// TODO(annasong): implement
func (lbg *localizeBuiltinGenerators) Filter(generators []*yaml.RNode) ([]*yaml.RNode, error) {
	return generators, nil
}

// localizeBuiltinTransformers localizes built-in transformers with file paths.
type localizeBuiltinTransformers struct {
}

var _ kio.Filter = &localizeBuiltinTransformers{}

// Filter localizes the built-in transformers with file paths. Filter returns an error if
// transformers contains a resource that is not a built-in transformer, cannot contain a file path,
// or is not localizable.
// TODO(annasong): implement
func (lbt *localizeBuiltinTransformers) Filter(transformers []*yaml.RNode) ([]*yaml.RNode, error) {
	return transformers, nil
}
