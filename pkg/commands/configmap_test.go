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

package commands

import (
	"testing"

	"k8s.io/kubectl/pkg/kustomize/types"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

func TestNewAddConfigMapIsNotNil(t *testing.T) {
	if newCmdAddConfigMap(nil, fs.MakeFakeFS()) == nil {
		t.Fatal("newCmdAddConfigMap shouldn't be nil")
	}
}

func TestGetOrCreateConfigMap(t *testing.T) {
	cmName := "test-config-name"

	kustomization := &types.Kustomization{
		NamePrefix: "test-name-prefix",
	}

	if len(kustomization.ConfigMapGenerator) != 0 {
		t.Fatal("Initial kustomization should not have any configmaps")
	}
	cm := getOrCreateConfigMap(kustomization, cmName)

	if cm == nil {
		t.Fatalf("ConfigMap should always be non-nil")
	}

	if len(kustomization.ConfigMapGenerator) != 1 {
		t.Fatalf("Kustomization should have newly created configmap")
	}

	if &kustomization.ConfigMapGenerator[len(kustomization.ConfigMapGenerator)-1] != cm {
		t.Fatalf("Pointer address for newly inserted configmap should be same")
	}

	existingCM := getOrCreateConfigMap(kustomization, cmName)

	if existingCM != cm {
		t.Fatalf("should have returned an existing cm with name: %v", cmName)
	}

	if len(kustomization.ConfigMapGenerator) != 1 {
		t.Fatalf("Should not insert configmap for an existing name: %v", cmName)
	}
}

func TestMergeData_LiteralSources(t *testing.T) {
	ds := &types.DataSources{}

	err := mergeData(ds, dataConfig{LiteralSources: []string{"k1=v1"}})
	if err != nil {
		t.Fatalf("Merge initial literal source should not return error")
	}

	if len(ds.LiteralSources) != 1 {
		t.Fatalf("Initial literal source should have been added")
	}

	err = mergeData(ds, dataConfig{LiteralSources: []string{"k2=v2"}})
	if err != nil {
		t.Fatalf("Merge second literal source should not return error")
	}

	if len(ds.LiteralSources) != 2 {
		t.Fatalf("Second literal source should have been added")
	}
}

func TestMergeData_FileSources(t *testing.T) {
	ds := &types.DataSources{}

	err := mergeData(ds, dataConfig{FileSources: []string{"file1"}})
	if err != nil {
		t.Fatalf("Merge initial file source should not return error")
	}

	if len(ds.FileSources) != 1 {
		t.Fatalf("Initial file source should have been added")
	}

	err = mergeData(ds, dataConfig{FileSources: []string{"file2"}})
	if err != nil {
		t.Fatalf("Merge second file source should not return error")
	}

	if len(ds.FileSources) != 2 {
		t.Fatalf("Second file source should have been added")
	}
}

func TestMergeData_EnvSource(t *testing.T) {
	envFileName := "env1"
	envFileName2 := "env2"
	ds := &types.DataSources{}

	err := mergeData(ds, dataConfig{EnvFileSource: envFileName})
	if err != nil {
		t.Fatalf("Merge initial env source should not return error")
	}

	if ds.EnvSource != envFileName {
		t.Fatalf("Initial env source filename should have been added")
	}

	err = mergeData(ds, dataConfig{EnvFileSource: envFileName2})
	if err == nil {
		t.Fatalf("Updating env source should return an error")
	}
}
