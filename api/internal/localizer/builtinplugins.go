// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinhelpers"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
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
//
// Note that the actual localization has not been implemented yet.
func (lbt *localizeBuiltinTransformers) Filter(transformers []*yaml.RNode) ([]*yaml.RNode, error) {
	doNothingFn := func(_ *yaml.RNode) error { return nil }
	filterParams := map[builtinhelpers.BuiltinPluginType]fieldspec.Filter{
		// TODO(annasong): replace doNothingFn with actual localize function
		builtinhelpers.PatchTransformer: {
			FieldSpec: types.FieldSpec{Path: "path"},
			SetValue:  doNothingFn,
		},
		builtinhelpers.PatchJson6902Transformer: {
			FieldSpec: types.FieldSpec{Path: "path"},
			SetValue:  doNothingFn,
		},
		builtinhelpers.PatchStrategicMergeTransformer: {
			FieldSpec: types.FieldSpec{Path: "paths[]"},
			SetValue:  doNothingFn,
		},
		builtinhelpers.ReplacementTransformer: {
			FieldSpec: types.FieldSpec{Path: "replacements[]/path"},
			SetValue:  doNothingFn,
		},
	}
	newTransformers, err := kio.FilterAll(yaml.FilterFunc(func(transformer *yaml.RNode) (*yaml.RNode, error) {
		printed, err := transformer.String()
		if err != nil {
			printed = fmt.Sprintf("<%s>", err)
		}

		apiVersion := transformer.GetApiVersion()
		if apiVersion != konfig.BuiltinPluginApiVersion {
			return nil, errors.Errorf("apiVersion %q of transformer %q is not built-in", apiVersion, printed)
		}
		kind := transformer.GetKind()
		builtinType := builtinhelpers.GetBuiltinPluginType(kind)
		if filter, exists := filterParams[builtinType]; exists {
			newTransformer, err := filter.Filter(transformer)
			return newTransformer, errors.Wrap(err)
		}
		return nil, errors.Errorf("built-in transformer %q of kind %q does not have file path", printed, kind)
	})).Filter(transformers)
	return newTransformers, errors.Wrap(err)
}
