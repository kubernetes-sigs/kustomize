// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/plugin/builtin"
)

//go:generate stringer -type=BuiltinPluginType
type BuiltinPluginType int

const (
	Unknown BuiltinPluginType = iota
	SecretGenerator
	ConfigMapGenerator
	ReplicaCountTransformer
	NamespaceTransformer
	PatchJson6902Transformer
	PatchStrategicMergeTransformer
	PatchTransformer
	LabelTransformer
	AnnotationsTransformer
	PrefixSuffixTransformer
	ImageTagTransformer
	HashTransformer
	InventoryTransformer
	LegacyOrderTransformer
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
	SecretGenerator:    builtin.NewSecretGeneratorPlugin,
	ConfigMapGenerator: builtin.NewConfigMapGeneratorPlugin,
}

var TransformerFactories = map[BuiltinPluginType]func() resmap.TransformerPlugin{
	NamespaceTransformer:           builtin.NewNamespaceTransformerPlugin,
	ReplicaCountTransformer:        builtin.NewReplicaCountTransformerPlugin,
	PatchJson6902Transformer:       builtin.NewPatchJson6902TransformerPlugin,
	PatchStrategicMergeTransformer: builtin.NewPatchStrategicMergeTransformerPlugin,
	PatchTransformer:               builtin.NewPatchTransformerPlugin,
	LabelTransformer:               builtin.NewLabelTransformerPlugin,
	AnnotationsTransformer:         builtin.NewAnnotationsTransformerPlugin,
	PrefixSuffixTransformer:        builtin.NewPrefixSuffixTransformerPlugin,
	ImageTagTransformer:            builtin.NewImageTagTransformerPlugin,
	HashTransformer:                builtin.NewHashTransformerPlugin,
	InventoryTransformer:           builtin.NewInventoryTransformerPlugin,
	LegacyOrderTransformer:         builtin.NewLegacyOrderTransformerPlugin,
}
