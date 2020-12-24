// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
)

// Harness manages a test environment.
type Harness struct {
	t    *testing.T
	fSys filesys.FileSystem
}

func MakeHarness(t *testing.T) Harness {
	return MakeHarnessWithFs(t, filesys.MakeFsInMemory())
}

func MakeHarnessWithFs(
	t *testing.T, fSys filesys.FileSystem) Harness {
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
	th.fSys.WriteFile(
		filepath.Join(
			path,
			konfig.DefaultKustomizationFileName()), []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`+content))
}

func (th Harness) WriteC(path string, content string) {
	th.fSys.WriteFile(
		filepath.Join(
			path,
			konfig.DefaultKustomizationFileName()), []byte(`
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
`+content))
}

func (th Harness) WriteF(path string, content string) {
	th.fSys.WriteFile(path, []byte(content))
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
	pc, err := konfig.EnabledPluginConfig(types.BploLoadFromFileSys)
	if err != nil {
		if strings.Contains(err.Error(), "unable to find plugin root") {
			th.t.Log(
				"Tests that want to run with plugins enabled must be " +
					"bookended by calls to MakeEnhancedHarness(), Reset().")
		}
		th.t.Fatal(err)
	}
	o := *krusty.MakeDefaultOptions()
	o.PluginConfig = pc
	return o
}

// Run, failing on error.
func (th Harness) Run(path string, o krusty.Options) resmap.ResMap {
	m, err := krusty.MakeKustomizer(th.fSys, &o).Run(path)
	if err != nil {
		th.t.Fatal(err)
	}
	return m
}

// Run, failing if there is no error.
func (th Harness) RunWithErr(path string, o krusty.Options) error {
	_, err := krusty.MakeKustomizer(th.fSys, &o).Run(path)
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

func (th Harness) AssertActualEqualsExpectedWithTweak(
	m resmap.ResMap, tweaker func([]byte) []byte, expected string) {
	assertActualEqualsExpectedWithTweak(th, m, tweaker, expected)
}
