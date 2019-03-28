/*
Copyright 2019 The Kubernetes Authors.

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

package configmapandsecret

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kv"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/types"
)

func TestKeyValuesFromFileSources(t *testing.T) {
	tests := []struct {
		description string
		sources     []string
		expected    []kv.Pair
	}{
		{
			description: "create kvs from file sources",
			sources:     []string{"files/app-init.ini"},
			expected: []kv.Pair{
				{
					Key:   "app-init.ini",
					Value: "FOO=bar",
				},
			},
		},
	}

	fSys := fs.MakeFakeFS()
	fSys.WriteFile("/files/app-init.ini", []byte("FOO=bar"))
	ldr := loader.NewFileLoaderAtRoot(fSys)
	reg := plugin.NewRegistry(ldr)
	bf := baseFactory{loader.NewFileLoaderAtRoot(fSys), nil, reg}
	for _, tc := range tests {
		kvs, err := bf.keyValuesFromFileSources(tc.sources)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(kvs, tc.expected) {
			t.Fatalf("in testcase: %q updated:\n%#v\ndoesn't match expected:\n%#v\n", tc.description, kvs, tc.expected)
		}
	}
}

func TestKeyValuesFromPlugins(t *testing.T) {
	tests := []struct {
		description string
		sources     []types.KVSource
		expected    []kv.Pair
	}{
		{
			description: "Create kv.Pairs from builtin literals plugin",
			sources: []types.KVSource{
				{
					PluginType: "builtin",
					Name:       "literals",
					Args:       []string{"FOO=bar", "BAR=baz"},
				},
			},
			expected: []kv.Pair{
				{
					Key:   "FOO",
					Value: "bar",
				},
				{
					Key:   "BAR",
					Value: "baz",
				},
			},
		},
		{
			description: "Create kv.Pairs from builtin files plugin",
			sources: []types.KVSource{
				{
					PluginType: "builtin",
					Name:       "files",
					Args:       []string{"files/app-init.ini"},
				},
			},
			expected: []kv.Pair{
				{
					Key:   "app-init.ini",
					Value: "FOO=bar",
				},
			},
		},
		{
			description: "Create kv.Pairs from builtin envfiles plugin",
			sources: []types.KVSource{
				{
					PluginType: "builtin",
					Name:       "envfiles",
					Args:       []string{"files/app-init.ini"},
				},
			},
			expected: []kv.Pair{
				{
					Key:   "FOO",
					Value: "bar",
				},
			},
		},
	}

	fSys := fs.MakeFakeFS()
	fSys.WriteFile("/files/app-init.ini", []byte("FOO=bar"))
	ldr := loader.NewFileLoaderAtRoot(fSys)
	reg := plugin.NewRegistry(ldr)
	bf := baseFactory{ldr, nil, reg}

	for _, tc := range tests {
		kvs, err := bf.keyValuesFromPlugins(tc.sources)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(kvs, tc.expected) {
			t.Fatalf("in testcase: %q updated:\n%#v\ndoesn't match expected:\n%#v\n", tc.description, kvs, tc.expected)
		}
	}
}
