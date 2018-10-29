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

func TestMakeDefaultTransformerConfig(t *testing.T) {
	// Confirm default can be made without fatal error inside call.
	l, _, _ := makeFakeLoaderAndOutput()
	_ = NewFactory(l).DefaultConfig()
}

func makeFakeLoaderAndOutput() (ifc.Loader, *TransformerConfig, *TransformerConfig) {
	transformerConfig := `
namePrefix:
- path: nameprefix/path
  kind: SomeKind
`
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile("/transformerconfig/test/config.yaml", []byte(transformerConfig))
	ldr := loader.NewFileLoaderAtRoot(fakeFS)
	expected := &TransformerConfig{
		NamePrefix: []FieldSpec{
			{
				Gvk:  gvk.Gvk{Kind: "SomeKind"},
				Path: "nameprefix/path",
			},
		},
	}
	return ldr, expected, NewFactory(ldr).DefaultConfig()
}

func TestMakeTransformerConfigFromFiles(t *testing.T) {
	ldr, expected, _ := makeFakeLoaderAndOutput()
	tcfg, err := NewFactory(ldr).FromFiles([]string{"/transformerconfig/test/config.yaml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(tcfg, expected) {
		t.Fatalf("expected %v\n but go6t %v\n", expected, tcfg)
	}
}
