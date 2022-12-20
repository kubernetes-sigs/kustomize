// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
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

	// locFn is used by localizeNode to set the localized path on the plugin.
	locFn func(string) (string, error)
}

var _ kio.Filter = &localizeBuiltinPlugins{}

// Filter localizes the built-in plugins with file paths.
func (lbp *localizeBuiltinPlugins) Filter(plugins []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, singlePlugin := range plugins {
		err := singlePlugin.PipeE(fsslice.Filter{
			FsSlice: types.FsSlice{
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.PatchTransformer.String()},
					Path: "path",
				},
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.PatchJson6902Transformer.String()},
					Path: "path",
				},
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.ReplacementTransformer.String()},
					Path: "replacements/path",
				},
			},
			SetValue: func(node *yaml.RNode) error {
				lbp.locFn = lbp.lc.localizeFile
				return lbp.localizeNode(node)
			},
		}, fieldspec.Filter{
			FieldSpec: types.FieldSpec{
				Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.PatchStrategicMergeTransformer.String()},
				Path: "paths",
			},
			SetValue: func(node *yaml.RNode) error {
				lbp.locFn = lbp.lc.localizeResource
				return errors.Wrap(node.VisitElements(lbp.localizeNode))
			},
		})
		// TODO(annasong): localize ConfigMapGenerator, SecretGenerator
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	return plugins, nil
}

// localizeNode sets the scalar node to its value localized by locFn.
func (lbp *localizeBuiltinPlugins) localizeNode(node *yaml.RNode) error {
	localizedPath, err := lbp.locFn(node.YNode().Value)
	if err != nil {
		return err
	}
	if localizedPath != "" {
		err = filtersutil.SetScalar(localizedPath)(node)
	}
	return err
}
