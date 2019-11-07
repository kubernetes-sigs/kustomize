// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinconfig_test

import (
	"reflect"
	"testing"

	. "sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
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
		FieldSpecs: []types.FieldSpec{
			{
				Gvk: resid.Gvk{
					Kind: "KindB",
				},
				Path:               "path/to/a/field",
				CreateIfNotPresent: false,
			},
		},
	}

	err := cfg.AddNamereferenceFieldSpec(nbrs)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(cfg.NameReference) != 1 {
		t.Fatal("failed to add namereference FieldSpec")
	}
}

func TestAddFieldSpecs(t *testing.T) {
	cfg := &TransformerConfig{}

	fieldSpec := types.FieldSpec{
		Gvk:                resid.Gvk{Group: "GroupA", Kind: "KindB"},
		Path:               "path/to/a/field",
		CreateIfNotPresent: true,
	}

	err := cfg.AddPrefixFieldSpec(fieldSpec)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(cfg.NamePrefix) != 1 {
		t.Fatalf("failed to add nameprefix FieldSpec")
	}
	err = cfg.AddSuffixFieldSpec(fieldSpec)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(cfg.NameSuffix) != 1 {
		t.Fatalf("failed to add namesuffix FieldSpec")
	}
	err = cfg.AddLabelFieldSpec(fieldSpec)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(cfg.CommonLabels) != 1 {
		t.Fatalf("failed to add nameprefix FieldSpec")
	}
	err = cfg.AddAnnotationFieldSpec(fieldSpec)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(cfg.CommonAnnotations) != 1 {
		t.Fatalf("failed to add nameprefix FieldSpec")
	}
}

func TestMerge(t *testing.T) {
	nameReference := []NameBackReferences{
		{
			Gvk: resid.Gvk{
				Kind: "KindA",
			},
			FieldSpecs: []types.FieldSpec{
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
			FieldSpecs: []types.FieldSpec{
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
	cfga.AddNamereferenceFieldSpec(nameReference[0])
	cfga.AddPrefixFieldSpec(fieldSpecs[0])
	cfga.AddSuffixFieldSpec(fieldSpecs[0])

	cfgb := &TransformerConfig{}
	cfgb.AddNamereferenceFieldSpec(nameReference[1])
	cfgb.AddPrefixFieldSpec(fieldSpecs[1])
	cfga.AddSuffixFieldSpec(fieldSpecs[1])

	actual, err := cfga.Merge(cfgb)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if len(actual.NamePrefix) != 2 {
		t.Fatal("merge failed for namePrefix FieldSpec")
	}

	if len(actual.NameSuffix) != 2 {
		t.Fatal("merge failed for nameSuffix FieldSpec")
	}

	if len(actual.NameReference) != 1 {
		t.Fatal("merge failed for namereference FieldSpec")
	}

	expected := &TransformerConfig{}
	expected.AddNamereferenceFieldSpec(nameReference[0])
	expected.AddNamereferenceFieldSpec(nameReference[1])
	expected.AddPrefixFieldSpec(fieldSpecs[0])
	expected.AddPrefixFieldSpec(fieldSpecs[1])
	expected.AddSuffixFieldSpec(fieldSpecs[0])
	expected.AddSuffixFieldSpec(fieldSpecs[1])

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected: %v\n but got: %v\n", expected, actual)
	}

	actual, err = cfga.Merge(nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(actual, cfga) {
		t.Fatalf("expected: %v\n but got: %v\n", cfga, actual)
	}
}
