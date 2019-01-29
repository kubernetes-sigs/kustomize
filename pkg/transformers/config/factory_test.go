/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"reflect"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/loader"
	"testing"
)

func TestMakeDefaultConfig(t *testing.T) {
	// Confirm default can be made without fatal error inside call.
	_ = MakeDefaultConfig()
}

func makeTestLoader(path, content string) ifc.Loader {
	fSys := fs.MakeFakeFS()
	fSys.WriteFile(path, []byte(content))
	return loader.NewFileLoaderAtRoot(fSys)
}

func TestFromFiles(t *testing.T) {
	path := "/transformerconfig/test/config.yaml"
	ldr := makeTestLoader(path, `
namePrefix:
- path: nameprefix/path
  kind: SomeKind
`)
	tcfg, err := NewFactory(ldr).FromFiles([]string{"transformerconfig/test/config.yaml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := &TransformerConfig{
		NamePrefix: []FieldSpec{
			{
				Gvk:  gvk.Gvk{Kind: "SomeKind"},
				Path: "nameprefix/path",
			},
		},
	}
	if !reflect.DeepEqual(tcfg, expected) {
		t.Fatalf("expected %v\n but go6t %v\n", expected, tcfg)
	}
}
