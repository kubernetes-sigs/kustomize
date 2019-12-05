// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinconfig

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
)

func TestLoadDefaultConfigsFromFiles(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile("config.yaml", []byte(`
namePrefix:
- path: nameprefix/path
  kind: SomeKind
`))
	if err != nil {
		t.Fatal(err)
	}
	ldr, err := loader.NewLoader(
		loader.RestrictionRootOnly, filesys.Separator, fSys)
	if err != nil {
		t.Fatal(err)
	}
	tCfg, err := loadDefaultConfig(ldr, []string{"config.yaml"})
	if err != nil {
		t.Fatal(err)
	}
	expected := &TransformerConfig{
		NamePrefix: []types.FieldSpec{
			{
				Gvk:  resid.Gvk{Kind: "SomeKind"},
				Path: "nameprefix/path",
			},
		},
	}
	if !reflect.DeepEqual(tCfg, expected) {
		t.Fatalf("expected %v\n but go6t %v\n", expected, tCfg)
	}
}
