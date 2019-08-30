// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/v3/internal/loadertest"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
)

func TestMakeDefaultConfig(t *testing.T) {
	// Confirm default can be made without fatal error inside call.
	_ = MakeDefaultConfig()
}

func TestFromFiles(t *testing.T) {

	ldr := loadertest.NewFakeLoader("/app")
	ldr.AddFile("/app/config.yaml", []byte(`
namePrefix:
- path: nameprefix/path
  kind: SomeKind
`))
	emptycfg := &TransformerConfig{}
	tcfg, err := NewFactory(ldr).FromFiles(emptycfg, []string{"/app/config.yaml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := &TransformerConfig{
		NamePrefix: []FieldSpecConfig{{
			FieldSpec: FieldSpec{
				Gvk:  gvk.Gvk{Kind: "SomeKind"},
				Path: "nameprefix/path",
			},
		}},
	}
	if !reflect.DeepEqual(tcfg, expected) {
		t.Fatalf("expected %v\n but got %v\n", expected, tcfg)
	}
}

func TestMakeTransformerConfig(t *testing.T) {

	ldr := loadertest.NewFakeLoader("/app")
	ldr.AddFile("/app/mycrdonly.yaml", []byte(`
namePrefix:
- path: metadata/name
  behavior: remove
- path: metadata/name
  kind: MyCRD
  behavior: add
`))
	tcfg, err := MakeTransformerConfig(ldr, []string{"/app/mycrdonly.yaml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := &TransformerConfig{
		NamePrefix: []FieldSpecConfig{
			{
				FieldSpec: FieldSpec{
					Gvk:                gvk.Gvk{Kind: "CustomResourceDefinition"},
					Path:               "metadata/name",
					SkipTransformation: true,
				},
			},
			{
				FieldSpec: FieldSpec{
					Gvk:  gvk.Gvk{Kind: "MyCRD"},
					Path: "metadata/name",
				},
			},
		},
	}
	if !reflect.DeepEqual(tcfg.NamePrefixFieldSpecs(), expected.NamePrefixFieldSpecs()) {
		t.Fatalf("expected %v\n but got %v\n", expected.NamePrefixFieldSpecs(), tcfg.NamePrefixFieldSpecs())
	}
}
