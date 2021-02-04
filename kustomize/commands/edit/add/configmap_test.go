// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/loader"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestNewAddConfigMapIsNotNil(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	if newCmdAddConfigMap(
		fSys,
		kv.NewLoader(
			loader.NewFileLoaderAtCwd(fSys),
			valtest_test.MakeFakeValidator()),
		nil) == nil {
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
	args := findOrMakeConfigMapArgs(kustomization, cmName)

	if args == nil {
		t.Fatalf("args should always be non-nil")
	}

	if len(kustomization.ConfigMapGenerator) != 1 {
		t.Fatalf("Kustomization should have newly created configmap")
	}

	if &kustomization.ConfigMapGenerator[len(kustomization.ConfigMapGenerator)-1] != args {
		t.Fatalf("Pointer address for newly inserted configmap generator should be same")
	}

	args2 := findOrMakeConfigMapArgs(kustomization, cmName)

	if args2 != args {
		t.Fatalf("should have returned an existing args with name: %v", cmName)
	}

	if len(kustomization.ConfigMapGenerator) != 1 {
		t.Fatalf("Should not insert configmap for an existing name: %v", cmName)
	}
}

func TestMergeFlagsIntoConfigMapArgs_LiteralSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{LiteralSources: []string{"k1=v1"}})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{LiteralSources: []string{"k2=v2"}})
	if k.ConfigMapGenerator[0].LiteralSources[0] != "k1=v1" {
		t.Fatalf("expected v1")
	}
	if k.ConfigMapGenerator[0].LiteralSources[1] != "k2=v2" {
		t.Fatalf("expected v2")
	}
}

func TestMergeFlagsIntoConfigMapArgs_FileSources(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{FileSources: []string{"file1"}})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{FileSources: []string{"file2"}})
	if k.ConfigMapGenerator[0].FileSources[0] != "file1" {
		t.Fatalf("expected file1")
	}
	if k.ConfigMapGenerator[0].FileSources[1] != "file2" {
		t.Fatalf("expected file2")
	}
}

func TestMergeFlagsIntoConfigMapArgs_EnvSource(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{EnvFileSource: "env1"})
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		flagsAndArgs{EnvFileSource: "env2"})
	if k.ConfigMapGenerator[0].EnvSources[0] != "env1" {
		t.Fatalf("expected env1")
	}
	if k.ConfigMapGenerator[0].EnvSources[1] != "env2" {
		t.Fatalf("expected env2")
	}
}

func TestMergeFlagsIntoConfigMapArgs_Behavior(t *testing.T) {
	k := &types.Kustomization{}
	args := findOrMakeConfigMapArgs(k, "foo")

	createBehaviorFlags := flagsAndArgs{
		Behavior:      "create",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		createBehaviorFlags)
	if k.ConfigMapGenerator[0].Behavior != "create" {
		t.Fatalf("expected create")
	}

	mergeBehaviorFlags := flagsAndArgs{
		Behavior:      "merge",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		mergeBehaviorFlags)
	if k.ConfigMapGenerator[0].Behavior != "merge" {
		t.Fatalf("expected merge")
	}

	replaceBehaviorFlags := flagsAndArgs{
		Behavior:      "replace",
		EnvFileSource: "env1",
	}
	mergeFlagsIntoGeneratorArgs(
		&args.GeneratorArgs,
		replaceBehaviorFlags)
	if k.ConfigMapGenerator[0].Behavior != "replace" {
		t.Fatalf("expected replace")
	}
}
