// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinconfig_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func TestMakeDefaultConfig(t *testing.T) {
	// Confirm default can be made without fatal error inside call.
	_ = MakeDefaultConfig()
}

func TestAddNamereferenceFieldSpec(t *testing.T) {
	cfg := &TransformerConfig{}

	nbrs := NameBackReferences{
		Gvk: resid.Gvk{
			Kind: "KindA",
		},
		Referrers: []types.FieldSpec{
			{
				Gvk: resid.Gvk{
					Kind: "KindB",
				},
				Path:               "path/to/a/field",
				CreateIfNotPresent: false,
			},
		},
	}

	require.NoError(t, cfg.AddNamereferenceFieldSpec(nbrs))
	require.Len(t, cfg.NameReference, 1, "failed to add namereference FieldSpec")
}

func TestAddFieldSpecs(t *testing.T) {
	cfg := &TransformerConfig{}

	fieldSpec := types.FieldSpec{
		Gvk:                resid.Gvk{Group: "GroupA", Kind: "KindB"},
		Path:               "path/to/a/field",
		CreateIfNotPresent: true,
	}

	require.NoError(t, cfg.AddPrefixFieldSpec(fieldSpec))
	require.Len(t, cfg.NamePrefix, 1, "failed to add nameprefix FieldSpec")
	require.NoError(t, cfg.AddSuffixFieldSpec(fieldSpec))
	require.Len(t, cfg.NameSuffix, 1, "failed to add namesuffix FieldSpec")
	require.NoError(t, cfg.AddCommonLabelsFieldSpec(fieldSpec))
	require.Len(t, cfg.CommonLabels, 1, "failed to add labels FieldSpec")
	require.NoError(t, cfg.AddAnnotationFieldSpec(fieldSpec))
	require.Len(t, cfg.CommonAnnotations, 1, "failed to add nameprefix FieldSpec")
}

func TestMerge(t *testing.T) {
	nameReference := []NameBackReferences{
		{
			Gvk: resid.Gvk{
				Kind: "KindA",
			},
			Referrers: []types.FieldSpec{
				{
					Gvk: resid.Gvk{
						Kind: "KindB",
					},
					Path:               "path/to/a/field",
					CreateIfNotPresent: false,
				},
			},
		},
		{
			Gvk: resid.Gvk{
				Kind: "KindA",
			},
			Referrers: []types.FieldSpec{
				{
					Gvk: resid.Gvk{
						Kind: "KindC",
					},
					Path:               "path/to/a/field",
					CreateIfNotPresent: false,
				},
			},
		},
	}
	fieldSpecs := []types.FieldSpec{
		{
			Gvk:                resid.Gvk{Group: "GroupA", Kind: "KindB"},
			Path:               "path/to/a/field",
			CreateIfNotPresent: true,
		},
		{
			Gvk:                resid.Gvk{Group: "GroupA", Kind: "KindC"},
			Path:               "path/to/a/field",
			CreateIfNotPresent: true,
		},
	}
	cfga := &TransformerConfig{}
	require.NoError(t, cfga.AddNamereferenceFieldSpec(nameReference[0]))
	require.NoError(t, cfga.AddPrefixFieldSpec(fieldSpecs[0]))
	require.NoError(t, cfga.AddSuffixFieldSpec(fieldSpecs[0]))
	require.NoError(t, cfga.AddCommonLabelsFieldSpec(fieldSpecs[0]))
	require.NoError(t, cfga.AddLabelsFieldSpec(fieldSpecs[0]))

	cfgb := &TransformerConfig{}
	require.NoError(t, cfgb.AddNamereferenceFieldSpec(nameReference[1]))
	require.NoError(t, cfgb.AddPrefixFieldSpec(fieldSpecs[1]))
	require.NoError(t, cfgb.AddSuffixFieldSpec(fieldSpecs[1]))
	require.NoError(t, cfgb.AddCommonLabelsFieldSpec(fieldSpecs[1]))
	require.NoError(t, cfgb.AddLabelsFieldSpec(fieldSpecs[1]))

	actual, err := cfga.Merge(cfgb)
	require.NoError(t, err)
	require.Len(t, actual.NamePrefix, 2, "merge failed for namePrefix FieldSpec")
	require.Len(t, actual.NameSuffix, 2, "merge failed for nameSuffix FieldSpec")
	require.Len(t, actual.NameReference, 1, "merge failed for nameReference FieldSpec")
	require.Len(t, actual.Labels, 2, "merge failed for labels FieldSpec")
	require.Len(t, actual.CommonLabels, 2, "merge failed for commonLabels FieldSpec")

	expected := &TransformerConfig{}
	require.NoError(t, expected.AddNamereferenceFieldSpec(nameReference[0]))
	require.NoError(t, expected.AddNamereferenceFieldSpec(nameReference[1]))
	require.NoError(t, expected.AddPrefixFieldSpec(fieldSpecs[0]))
	require.NoError(t, expected.AddPrefixFieldSpec(fieldSpecs[1]))
	require.NoError(t, expected.AddSuffixFieldSpec(fieldSpecs[0]))
	require.NoError(t, expected.AddSuffixFieldSpec(fieldSpecs[1]))
	require.NoError(t, expected.AddCommonLabelsFieldSpec(fieldSpecs[0]))
	require.NoError(t, expected.AddCommonLabelsFieldSpec(fieldSpecs[1]))
	require.NoError(t, expected.AddLabelsFieldSpec(fieldSpecs[0]))
	require.NoError(t, expected.AddLabelsFieldSpec(fieldSpecs[1]))
	require.Equal(t, expected, actual)

	actual, err = cfga.Merge(nil)
	require.NoError(t, err)
	require.Equal(t, cfga, actual)
}

func TestMakeDefaultConfig_mutation(t *testing.T) {
	a := MakeDefaultConfig()

	// mutate
	a.NameReference[0].Kind = "mutated"
	a.NameReference = a.NameReference[:1]

	clean := MakeDefaultConfig()
	assert.NotEqualf(t, "mutated", clean.NameReference[0].Kind, "MakeDefaultConfig() did not return a clean copy: %+v", clean.NameReference)
}

func BenchmarkMakeDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MakeDefaultConfig()
	}
}
