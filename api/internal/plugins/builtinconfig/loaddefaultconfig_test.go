// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinconfig

import (
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/internal/loader"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/resid"
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

func TestLoadDefaultConfigsFromFilesWithMissingFields(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	filePathContainsTypo := "config_contains_typo.yaml"
	if err := fSys.WriteFile(filePathContainsTypo, []byte(`
namoPrefix:
- path: nameprefix/path
  kind: SomeKind
`)); err != nil {
		t.Fatal(err)
	}
	ldr, err := loader.NewLoader(
		loader.RestrictionRootOnly, filesys.Separator, fSys)
	if err != nil {
		t.Fatal(err)
	}
	errMsg := "error unmarshaling JSON: while decoding JSON: json: unknown field"
	_, err = loadDefaultConfig(ldr, []string{filePathContainsTypo})
	if err == nil {
		t.Fatalf("expected to fail unmarshal yaml, but got nil %s", filePathContainsTypo)
	}
	if !strings.Contains(err.Error(), errMsg) {
		t.Fatalf("expected error %s, but got %s", errMsg, err)
	}
}
