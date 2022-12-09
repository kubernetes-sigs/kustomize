// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinhelpers"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// localizeBuiltinPlugins localizes built-in plugins with file paths.
// Note that this excludes helm, which needs a repo.
type localizeBuiltinPlugins struct {
	lc *localizer
}

var _ kio.Filter = &localizeBuiltinPlugins{}

// Filter localizes the built-in plugins with file paths.
func (lbp *localizeBuiltinPlugins) Filter(plugins []*yaml.RNode) ([]*yaml.RNode, error) {
	newPlugins, err := kio.FilterAll(fsslice.Filter{
		FsSlice: types.FsSlice{
			types.FieldSpec{
				Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.PatchTransformer.String()},
				Path: "path",
			},
			types.FieldSpec{
				Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.PatchJson6902Transformer.String()},
				Path: "path",
			},
		},
		SetValue: lbp.evaluateField,
	}).Filter(plugins)

	// TODO(annasong): localize replacements, patchesStrategicMerge, configMapGenerator, secretGenerator

	return newPlugins, errors.Wrap(err)
}

func (lbp *localizeBuiltinPlugins) evaluateField(node *yaml.RNode) error {
	newPath, err := lbp.lc.localizeFile(node.YNode().Value)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to localize built-in plugin path")
	}
	return filtersutil.SetScalar(newPath)(node)
}
