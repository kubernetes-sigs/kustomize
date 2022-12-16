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
	localizedPlugins, err := kio.FilterAll(fsslice.Filter{
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
		SetValue: lbp.localizeNode,
	}).Filter(plugins)

	// TODO(annasong): localize ReplacementTransformer, PatchStrategicMergeTransformer, ConfigMapGenerator, SecretGenerator

	return localizedPlugins, errors.Wrap(err)
}

// localizeNode sets the scalar node to its value localized as a file path.
func (lbp *localizeBuiltinPlugins) localizeNode(node *yaml.RNode) error {
	localizedPath, err := lbp.lc.localizeFile(node.YNode().Value)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to localize built-in plugin path")
	}
	return filtersutil.SetScalar(localizedPath)(node)
}
