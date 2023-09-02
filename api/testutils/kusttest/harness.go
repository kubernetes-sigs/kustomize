// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/api/internal/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Harness manages a test environment.
type Harness struct {
	t    *testing.T
	fSys filesys.FileSystem
}

func MakeHarness(t *testing.T) Harness {
	t.Helper()
	return MakeHarnessWithFs(t, filesys.MakeFsInMemory())
}

func MakeHarnessWithFs(
	t *testing.T, fSys filesys.FileSystem) Harness {
	t.Helper()
	return Harness{
		t:    t,
		fSys: fSys,
	}
}

func (th Harness) GetT() *testing.T {
	return th.t
}

func (th Harness) GetFSys() filesys.FileSystem {
	return th.fSys
}

func (th Harness) WriteK(path string, content string) {
	err := th.fSys.WriteFile(
		filepath.Join(
			path,
			konfig.DefaultKustomizationFileName()), []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`+content))
	if err != nil {
		th.t.Fatalf("unexpected error while writing Kustomization to %s: %v", path, err)
	}
}

func (th Harness) WriteC(path string, content string) {
	err := th.fSys.WriteFile(
		filepath.Join(
			path,
			konfig.DefaultKustomizationFileName()), []byte(`
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
`+content))
	if err != nil {
		th.t.Fatalf("unexpected error while writing Component to %s: %v", path, err)
	}
}

func (th Harness) WriteF(path string, content string) {
	err := th.fSys.WriteFile(path, []byte(content))
	if err != nil {
		th.t.Fatalf("unexpected error while writing file to %s: %v", path, err)
	}
}

func (th Harness) MakeDefaultOptions() krusty.Options {
	return th.MakeOptionsPluginsDisabled()
}

// This has no impact on Builtin plugins, as they are always enabled.
func (th Harness) MakeOptionsPluginsDisabled() krusty.Options {
	return *krusty.MakeDefaultOptions()
}

// Enables use of non-builtin plugins.
func (th Harness) MakeOptionsPluginsEnabled() krusty.Options {
	pc := types.EnabledPluginConfig(types.BploLoadFromFileSys)
	o := *krusty.MakeDefaultOptions()
	o.PluginConfig = pc
	return o
}

// Run, failing on error.
func (th Harness) Run(path string, o krusty.Options) resmap.ResMap {
	m, err := krusty.MakeKustomizer(&o).Run(th.fSys, path)
	if err != nil {
		th.t.Fatal(err)
	}
	return m
}

// Run, failing if there is no error.
func (th Harness) RunWithErr(path string, o krusty.Options) error {
	_, err := krusty.MakeKustomizer(&o).Run(th.fSys, path)
	if err == nil {
		th.t.Fatalf("expected error")
	}
	return err
}

func (th Harness) WriteLegacyConfigs(fName string) {
	m := builtinpluginconsts.GetDefaultFieldSpecsAsMap()
	var content []byte
	for _, tCfg := range m {
		content = append(content, []byte(tCfg)...)
	}
	err := th.fSys.WriteFile(fName, content)
	if err != nil {
		th.t.Fatalf("unable to add file %s", fName)
	}
}

func (th Harness) AssertActualEqualsExpected(
	m resmap.ResMap, expected string) {
	th.AssertActualEqualsExpectedWithTweak(m, nil, expected)
}

func (th Harness) AssertActualEqualsExpectedNoIdAnnotations(m resmap.ResMap, expected string) {
	m.RemoveBuildAnnotations()
	th.AssertActualEqualsExpectedWithTweak(m, nil, expected)
}

func (th Harness) AssertActualEqualsExpectedWithTweak(
	m resmap.ResMap, tweaker func([]byte) []byte, expected string) {
	assertActualEqualsExpectedWithTweak(th, m, tweaker, expected)
}
