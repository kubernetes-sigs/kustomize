// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinhelpers

import (
	"sigs.k8s.io/kustomize/v3/api/resmap"
	"sigs.k8s.io/kustomize/v3/plugin/builtin"
)

//go:generate stringer -type=BuiltinPluginType
type BuiltinPluginType int

const (
	Unknown BuiltinPluginType = iota
	AnnotationsTransformer
	ConfigMapGenerator
	HashTransformer
	ImageTagTransformer
	InventoryTransformer
	LabelTransformer
	LegacyOrderTransformer
	NamespaceTransformer
	PatchJson6902Transformer
	PatchStrategicMergeTransformer
	PatchTransformer
	PrefixSuffixTransformer
	ReplicaCountTransformer
	SecretGenerator
)

var stringToBuiltinPluginTypeMap map[string]BuiltinPluginType

func init() {
	stringToBuiltinPluginTypeMap = makeStringToBuiltinPluginTypeMap()
}

func makeStringToBuiltinPluginTypeMap() (result map[string]BuiltinPluginType) {
	result = make(map[string]BuiltinPluginType, 23)
	for k := range GeneratorFactories {
		result[k.String()] = k
	}
	for k := range TransformerFactories {
		result[k.String()] = k
	}
	return
}

func GetBuiltinPluginType(n string) BuiltinPluginType {
	result, ok := stringToBuiltinPluginTypeMap[n]
	if ok {
		return result
	}
	return Unknown
}

var GeneratorFactories = map[BuiltinPluginType]func() resmap.GeneratorPlugin{
	ConfigMapGenerator: builtin.NewConfigMapGeneratorPlugin,
	SecretGenerator:    builtin.NewSecretGeneratorPlugin,
}

var TransformerFactories = map[BuiltinPluginType]func() resmap.TransformerPlugin{
	AnnotationsTransformer:         builtin.NewAnnotationsTransformerPlugin,
	HashTransformer:                builtin.NewHashTransformerPlugin,
	ImageTagTransformer:            builtin.NewImageTagTransformerPlugin,
	InventoryTransformer:           builtin.NewInventoryTransformerPlugin,
	LabelTransformer:               builtin.NewLabelTransformerPlugin,
	LegacyOrderTransformer:         builtin.NewLegacyOrderTransformerPlugin,
	NamespaceTransformer:           builtin.NewNamespaceTransformerPlugin,
	PatchJson6902Transformer:       builtin.NewPatchJson6902TransformerPlugin,
	PatchStrategicMergeTransformer: builtin.NewPatchStrategicMergeTransformerPlugin,
	PatchTransformer:               builtin.NewPatchTransformerPlugin,
	PrefixSuffixTransformer:        builtin.NewPrefixSuffixTransformerPlugin,
	ReplicaCountTransformer:        builtin.NewReplicaCountTransformerPlugin,
}
