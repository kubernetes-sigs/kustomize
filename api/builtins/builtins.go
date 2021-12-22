// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Deprecated: Package api/builtins will not be available in API v1.
package builtins

import (
	internal "sigs.k8s.io/kustomize/api/internal/builtins"
)

type (
	AnnotationsTransformerPlugin         = internal.AnnotationsTransformerPlugin
	ConfigMapGeneratorPlugin             = internal.ConfigMapGeneratorPlugin
	HashTransformerPlugin                = internal.HashTransformerPlugin
	HelmChartInflationGeneratorPlugin    = internal.HelmChartInflationGeneratorPlugin
	IAMPolicyGeneratorPlugin             = internal.IAMPolicyGeneratorPlugin
	ImageTagTransformerPlugin            = internal.ImageTagTransformerPlugin
	LabelTransformerPlugin               = internal.LabelTransformerPlugin
	LegacyOrderTransformerPlugin         = internal.LegacyOrderTransformerPlugin
	NamespaceTransformerPlugin           = internal.NamespaceTransformerPlugin
	PatchJson6902TransformerPlugin       = internal.PatchJson6902TransformerPlugin
	PatchStrategicMergeTransformerPlugin = internal.PatchStrategicMergeTransformerPlugin
	PatchTransformerPlugin               = internal.PatchTransformerPlugin
	PrefixTransformerPlugin              = internal.PrefixTransformerPlugin
	SuffixTransformerPlugin              = internal.SuffixTransformerPlugin
	ReplacementTransformerPlugin         = internal.ReplacementTransformerPlugin
	ReplicaCountTransformerPlugin        = internal.ReplicaCountTransformerPlugin
	SecretGeneratorPlugin                = internal.SecretGeneratorPlugin
	ValueAddTransformerPlugin            = internal.ValueAddTransformerPlugin
)

var (
	NewAnnotationsTransformerPlugin         = internal.NewAnnotationsTransformerPlugin
	NewConfigMapGeneratorPlugin             = internal.NewConfigMapGeneratorPlugin
	NewHashTransformerPlugin                = internal.NewHashTransformerPlugin
	NewHelmChartInflationGeneratorPlugin    = internal.NewHelmChartInflationGeneratorPlugin
	NewIAMPolicyGeneratorPlugin             = internal.NewIAMPolicyGeneratorPlugin
	NewImageTagTransformerPlugin            = internal.NewImageTagTransformerPlugin
	NewLabelTransformerPlugin               = internal.NewLabelTransformerPlugin
	NewLegacyOrderTransformerPlugin         = internal.NewLegacyOrderTransformerPlugin
	NewNamespaceTransformerPlugin           = internal.NewNamespaceTransformerPlugin
	NewPatchJson6902TransformerPlugin       = internal.NewPatchJson6902TransformerPlugin
	NewPatchStrategicMergeTransformerPlugin = internal.NewPatchStrategicMergeTransformerPlugin
	NewPatchTransformerPlugin               = internal.NewPatchTransformerPlugin
	NewPrefixTransformerPlugin              = internal.NewPrefixTransformerPlugin
	NewSuffixTransformerPlugin              = internal.NewSuffixTransformerPlugin
	NewReplacementTransformerPlugin         = internal.NewReplacementTransformerPlugin
	NewReplicaCountTransformerPlugin        = internal.NewReplicaCountTransformerPlugin
	NewSecretGeneratorPlugin                = internal.NewSecretGeneratorPlugin
	NewValueAddTransformerPlugin            = internal.NewValueAddTransformerPlugin
)
