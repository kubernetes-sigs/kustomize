/*
Copyright 2017 The Kubernetes Authors.

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

package add

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/types"
)

func TestNewAddConfigMapIsNotNil(t *testing.T) {
	if newCmdAddConfigMap(fs.MakeFakeFS(), nil) == nil {
		t.Fatal("newCmdAddConfigMap shouldn't be nil")
	}
}

func TestMakeConfigMapArgs(t *testing.T) {
	cmName := "test-config-name"

	kustomization := &types.Kustomization{
		NamePrefix: "test-name-prefix",
	}

	if len(kustomization.ConfigMapGenerator) != 0 {
		t.Fatal("Initial kustomization should not have any configmaps")
	}
	args := makeConfigMapArgs(kustomization, cmName)

	if args == nil {
		t.Fatalf("args should always be non-nil")
	}

	if len(kustomization.ConfigMapGenerator) != 1 {
		t.Fatalf("Kustomization should have newly created configmap")
	}

	if &kustomization.ConfigMapGenerator[len(kustomization.ConfigMapGenerator)-1] != args {
		t.Fatalf("Pointer address for newly inserted configmap generator should be same")
	}

	args2 := makeConfigMapArgs(kustomization, cmName)

	if args2 != args {
		t.Fatalf("should have returned an existing args with name: %v", cmName)
	}

	if len(kustomization.ConfigMapGenerator) != 1 {
		t.Fatalf("Should not insert configmap for an existing name: %v", cmName)
	}
}

func TestMergeFlagsIntoCmArgs_LiteralSources(t *testing.T) {
	kv := []types.KVSource{}

	mergeFlagsIntoCmArgs(&kv, flagsAndArgs{LiteralSources: []string{"k1=v1"}})

	if len(kv) != 1 {
		t.Fatalf("Initial literal source should have been added")
	}

	mergeFlagsIntoCmArgs(&kv, flagsAndArgs{LiteralSources: []string{"k2=v2"}})

	if len(kv) != 2 {
		t.Fatalf("Second literal source should have been added")
	}
}

func TestMergeFlagsIntoCmArgs_FileSources(t *testing.T) {
	kv := []types.KVSource{}

	mergeFlagsIntoCmArgs(&kv, flagsAndArgs{FileSources: []string{"file1"}})

	if len(kv) != 1 {
		t.Fatalf("Initial file source should have been added")
	}

	mergeFlagsIntoCmArgs(&kv, flagsAndArgs{FileSources: []string{"file2"}})

	if len(kv) != 2 {
		t.Fatalf("Second file source should have been added")
	}
}

func TestMergeFlagsIntoCmArgs_EnvSource(t *testing.T) {
	envFileName := "env1"
	envFileName2 := "env2"
	kv := []types.KVSource{}

	mergeFlagsIntoCmArgs(&kv, flagsAndArgs{EnvFileSource: envFileName})

	if len(kv) != 1 {
		t.Fatalf("Initial env source should have been added")
	}

	mergeFlagsIntoCmArgs(&kv, flagsAndArgs{EnvFileSource: envFileName2})
	if len(kv) != 2 {
		t.Fatalf("Second env source should have been added")
	}
}
