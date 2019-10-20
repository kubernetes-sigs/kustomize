// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"sigs.k8s.io/kustomize/v3/api/builtinconfig"
	"sigs.k8s.io/kustomize/v3/api/plugins/builtins"
	"sigs.k8s.io/kustomize/v3/api/resmap"
	"sigs.k8s.io/kustomize/v3/api/types"
)

// Functions dedicated to configuring the builtin
// transformer and generator plugins using config data
// read from a kustomization file and from the
// config.TransformerConfig, whose data may be a
// mix of hardcoded values and data read from file.
//
// Non-builtin plugins will get their configuration
// from their own dedicated structs and YAML files.
//
// There are some loops in the functions below because
// the kustomization file would, say, allow someone to
// request multiple secrets be made, or run multiple
// image tag transforms.  In these cases, we'll need
// N plugin instances with differing configurations.

func (kt *KustTarget) configureBuiltinGenerators() (
	result []resmap.Generator, err error) {
	for _, bpt := range []builtins.BuiltinPluginType{
		builtins.ConfigMapGenerator,
		builtins.SecretGenerator,
	} {
		r, err := generatorConfigurators[bpt](
			kt, bpt, builtins.GeneratorFactories[bpt])
		if err != nil {
			return nil, err
		}
		result = append(result, r...)
	}
	return result, nil
}

func (kt *KustTarget) configureBuiltinTransformers(
	tc *builtinconfig.TransformerConfig) (
	result []resmap.Transformer, err error) {
	for _, bpt := range []builtins.BuiltinPluginType{
		builtins.PatchStrategicMergeTransformer,
		builtins.PatchTransformer,
		builtins.NamespaceTransformer,
		builtins.PrefixSuffixTransformer,
		builtins.LabelTransformer,
		builtins.AnnotationsTransformer,
		builtins.PatchJson6902Transformer,
		builtins.ReplicaCountTransformer,
		builtins.ImageTagTransformer,
	} {
		r, err := transformerConfigurators[bpt](
			kt, bpt, builtins.TransformerFactories[bpt], tc)
		if err != nil {
			return nil, err
		}
		result = append(result, r...)
	}
	return result, nil
}

type gFactory func() resmap.GeneratorPlugin

var generatorConfigurators = map[builtins.BuiltinPluginType]func(
	kt *KustTarget,
	bpt builtins.BuiltinPluginType,
	factory gFactory) (result []resmap.Generator, err error){
	builtins.SecretGenerator: func(kt *KustTarget, bpt builtins.BuiltinPluginType, f gFactory) (
		result []resmap.Generator, err error) {
		var c struct {
			types.GeneratorOptions
			types.SecretArgs
		}
		if kt.kustomization.GeneratorOptions != nil {
			c.GeneratorOptions = *kt.kustomization.GeneratorOptions
		}
		for _, args := range kt.kustomization.SecretGenerator {
			c.SecretArgs = args
			p := f()
			err := kt.configureBuiltinPlugin(p, c, bpt)
			if err != nil {
				return nil, err
			}
			result = append(result, p)
		}
		return
	},

	builtins.ConfigMapGenerator: func(kt *KustTarget, bpt builtins.BuiltinPluginType, f gFactory) (
		result []resmap.Generator, err error) {
		var c struct {
			types.GeneratorOptions
			types.ConfigMapArgs
		}
		if kt.kustomization.GeneratorOptions != nil {
			c.GeneratorOptions = *kt.kustomization.GeneratorOptions
		}
		for _, args := range kt.kustomization.ConfigMapGenerator {
			c.ConfigMapArgs = args
			p := f()
			err := kt.configureBuiltinPlugin(p, c, bpt)
			if err != nil {
				return nil, err
			}
			result = append(result, p)
		}
		return
	},
}

type tFactory func() resmap.TransformerPlugin

var transformerConfigurators = map[builtins.BuiltinPluginType]func(
	kt *KustTarget,
	bpt builtins.BuiltinPluginType,
	f tFactory,
	tc *builtinconfig.TransformerConfig) (result []resmap.Transformer, err error){
	builtins.NamespaceTransformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, tc *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		var c struct {
			types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
			FieldSpecs       []types.FieldSpec
		}
		c.Namespace = kt.kustomization.Namespace
		c.FieldSpecs = tc.NameSpace
		p := f()
		err = kt.configureBuiltinPlugin(p, c, bpt)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
		return
	},

	builtins.PatchJson6902Transformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, _ *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		var c struct {
			Target types.PatchTarget `json:"target,omitempty" yaml:"target,omitempty"`
			Path   string            `json:"path,omitempty" yaml:"path,omitempty"`
			JsonOp string            `json:"jsonOp,omitempty" yaml:"jsonOp,omitempty"`
		}
		for _, args := range kt.kustomization.PatchesJson6902 {
			c.Target = *args.Target
			c.Path = args.Path
			c.JsonOp = args.Patch
			p := f()
			err = kt.configureBuiltinPlugin(p, c, bpt)
			if err != nil {
				return nil, err
			}
			result = append(result, p)
		}
		return
	},
	builtins.PatchStrategicMergeTransformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, _ *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		if len(kt.kustomization.PatchesStrategicMerge) == 0 {
			return
		}
		var c struct {
			Paths   []types.PatchStrategicMerge `json:"paths,omitempty" yaml:"paths,omitempty"`
			Patches string                      `json:"patches,omitempty" yaml:"patches,omitempty"`
		}
		c.Paths = kt.kustomization.PatchesStrategicMerge
		p := f()
		err = kt.configureBuiltinPlugin(p, c, bpt)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
		return
	},
	builtins.PatchTransformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, _ *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		if len(kt.kustomization.Patches) == 0 {
			return
		}
		var c struct {
			Path   string          `json:"path,omitempty" yaml:"path,omitempty"`
			Patch  string          `json:"patch,omitempty" yaml:"patch,omitempty"`
			Target *types.Selector `json:"target,omitempty" yaml:"target,omitempty"`
		}
		for _, pc := range kt.kustomization.Patches {
			c.Target = pc.Target
			c.Patch = pc.Patch
			c.Path = pc.Path
			p := f()
			err = kt.configureBuiltinPlugin(p, c, bpt)
			if err != nil {
				return nil, err
			}
			result = append(result, p)
		}
		return
	},
	builtins.LabelTransformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, tc *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		var c struct {
			Labels     map[string]string
			FieldSpecs []types.FieldSpec
		}
		c.Labels = kt.kustomization.CommonLabels
		c.FieldSpecs = tc.CommonLabels
		p := f()
		err = kt.configureBuiltinPlugin(p, c, bpt)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
		return
	},
	builtins.AnnotationsTransformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, tc *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		var c struct {
			Annotations map[string]string
			FieldSpecs  []types.FieldSpec
		}
		c.Annotations = kt.kustomization.CommonAnnotations
		c.FieldSpecs = tc.CommonAnnotations
		p := f()
		err = kt.configureBuiltinPlugin(p, c, bpt)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
		return
	},
	builtins.PrefixSuffixTransformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, tc *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		var c struct {
			Prefix     string
			Suffix     string
			FieldSpecs []types.FieldSpec
		}
		c.Prefix = kt.kustomization.NamePrefix
		c.Suffix = kt.kustomization.NameSuffix
		c.FieldSpecs = tc.NamePrefix
		p := f()
		err = kt.configureBuiltinPlugin(p, c, bpt)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
		return
	},
	builtins.ImageTagTransformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, tc *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		var c struct {
			ImageTag   types.Image
			FieldSpecs []types.FieldSpec
		}
		for _, args := range kt.kustomization.Images {
			c.ImageTag = args
			c.FieldSpecs = tc.Images
			p := f()
			err = kt.configureBuiltinPlugin(p, c, bpt)
			if err != nil {
				return nil, err
			}
			result = append(result, p)
		}
		return
	},
	builtins.ReplicaCountTransformer: func(
		kt *KustTarget, bpt builtins.BuiltinPluginType, f tFactory, tc *builtinconfig.TransformerConfig) (
		result []resmap.Transformer, err error) {
		var c struct {
			Replica    types.Replica
			FieldSpecs []types.FieldSpec
		}
		for _, args := range kt.kustomization.Replicas {
			c.Replica = args
			c.FieldSpecs = tc.Replicas
			p := f()
			err = kt.configureBuiltinPlugin(p, c, bpt)
			if err != nil {
				return nil, err
			}
			result = append(result, p)
		}
		return
	},
}
