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

	// locPathFn is used by localizeNode to set the localized path on the plugin.
	locPathFn func(string) (string, error)
}

var _ kio.Filter = &localizeBuiltinPlugins{}

// Filter localizes the built-in plugins with file paths.
func (lbp *localizeBuiltinPlugins) Filter(plugins []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, singlePlugin := range plugins {
		err := singlePlugin.PipeE(fsslice.Filter{
			FsSlice: types.FsSlice{
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.ConfigMapGenerator.String()},
					Path: "env",
				},
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.ConfigMapGenerator.String()},
					Path: "envs",
				},
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.SecretGenerator.String()},
					Path: "env",
				},
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.SecretGenerator.String()},
					Path: "envs",
				},
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.HelmChartInflationGenerator.String()},
					Path: "valuesFile",
				},
				types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.HelmChartInflationGenerator.String()},
					Path: "additionalValuesFiles",
				},
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
				lbp.locPathFn = lbp.lc.localizeFile
				return lbp.localizeAll(node)
			},
		},
			fsslice.Filter{
				FsSlice: types.FsSlice{
					types.FieldSpec{
						Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.ConfigMapGenerator.String()},
						Path: "files",
					},
					types.FieldSpec{
						Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.SecretGenerator.String()},
						Path: "files",
					},
				},
				SetValue: func(node *yaml.RNode) error {
					lbp.locPathFn = lbp.lc.localizeFileSource
					return lbp.localizeAll(node)
				},
			},
			yaml.FilterFunc(func(node *yaml.RNode) (*yaml.RNode, error) {
				isHelm := node.GetApiVersion() == konfig.BuiltinPluginApiVersion &&
					node.GetKind() == builtinhelpers.HelmChartInflationGenerator.String()
				if !isHelm {
					return node, nil
				}
				home, err := node.Pipe(yaml.Lookup("chartHome"))
				if err != nil {
					return nil, errors.Wrap(err)
				}
				if home == nil {
					_, err = lbp.lc.copyChartHomeEntry("")
				} else {
					lbp.locPathFn = lbp.lc.copyChartHomeEntry
					err = lbp.localizeScalar(home)
				}
				return node, errors.WrapPrefixf(err, "plugin %s", resid.FromRNode(node))
			}),
			fieldspec.Filter{
				FieldSpec: types.FieldSpec{
					Gvk:  resid.Gvk{Version: konfig.BuiltinPluginApiVersion, Kind: builtinhelpers.PatchStrategicMergeTransformer.String()},
					Path: "paths",
				},
				SetValue: func(node *yaml.RNode) error {
					lbp.locPathFn = lbp.lc.localizeK8sResource
					return lbp.localizeAll(node)
				},
			})
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	return plugins, nil
}

// localizeAll sets each entry in node to its value localized by locPathFn.
// Node is a sequence or scalar value.
func (lbp *localizeBuiltinPlugins) localizeAll(node *yaml.RNode) error {
	// We rely on the build command to throw errors for nodes in
	// built-in plugins that are sequences when expected to be scalar,
	// and vice versa.
	//nolint: exhaustive
	switch node.YNode().Kind {
	case yaml.SequenceNode:
		return errors.Wrap(node.VisitElements(lbp.localizeScalar))
	case yaml.ScalarNode:
		return lbp.localizeScalar(node)
	default:
		return errors.Errorf("expected sequence or scalar node")
	}
}

// localizeScalar sets the scalar node to its value localized by locPathFn.
func (lbp *localizeBuiltinPlugins) localizeScalar(node *yaml.RNode) error {
	localizedPath, err := lbp.locPathFn(node.YNode().Value)
	if err != nil {
		return err
	}
	return filtersutil.SetScalar(localizedPath)(node)
}
