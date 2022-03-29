// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/internal/plugins/execplugin"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/plugins/utils"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestExecPluginConfig(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile("sed-input.txt", []byte(`
s/$FOO/foo/g
s/$BAR/bar baz/g
 \ \ \ 
`))
	require.NoError(t, err)
	ldr, err := fLdr.NewLoader(
		fLdr.RestrictionRootOnly, filesys.Separator, fSys, false, true)
	if err != nil {
		t.Fatal(err)
	}
	pvd := provider.NewDefaultDepProvider()
	rf := resmap.NewFactory(pvd.GetResourceFactory())
	pluginConfig := rf.RF().FromMap(
		map[string]interface{}{
			"apiVersion": "someteam.example.com/v1",
			"kind":       "SedTransformer",
			"metadata": map[string]interface{}{
				"name": "some-random-name",
			},
			"argsOneLiner": "one two 'foo bar'",
			"argsFromFile": "sed-input.txt",
		})

	pluginConfig.RemoveBuildAnnotations()
	pc := types.DisabledPluginConfig()
	loader := pLdr.NewLoader(pc, rf, fSys)
	pluginPath, err := loader.AbsolutePluginPath(pluginConfig.OrgId())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	p := NewExecPlugin(pluginPath)
	// Not checking to see if the plugin is executable,
	// because this test does not run it.
	// This tests only covers sending configuration
	// to the plugin wrapper object and confirming
	// that it's properly prepared for execution.

	yaml, err := pluginConfig.AsYAML()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = p.Config(
		resmap.NewPluginHelpers(ldr, pvd.GetFieldValidator(), rf, pc),
		yaml)
	require.NoError(t, err)

	expected := "someteam.example.com/v1/sedtransformer/SedTransformer"
	if !strings.HasSuffix(p.Path(), expected) {
		t.Fatalf("expected suffix '%s', got '%s'", expected, p.Path())
	}

	expected = `apiVersion: someteam.example.com/v1
argsFromFile: sed-input.txt
argsOneLiner: one two 'foo bar'
kind: SedTransformer
metadata:
  name: some-random-name
`
	if expected != string(p.Cfg()) {
		t.Fatalf("expected cfg '%s', got '%s'", expected, string(p.Cfg()))

	}
	if len(p.Args()) != 6 {
		t.Fatalf("unexpected arg len %d, %#v", len(p.Args()), p.Args())
	}
	if p.Args()[0] != "one" ||
		p.Args()[1] != "two" ||
		p.Args()[2] != "foo bar" ||
		p.Args()[3] != "s/$FOO/foo/g" ||
		p.Args()[4] != "s/$BAR/bar baz/g" ||
		p.Args()[5] != "\\ \\ \\ " {
		t.Fatalf("unexpected arg array: %#v", p.Args())
	}
}

func TestExecPlugin_ErrIfNotExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("always returns nil on Windows")
	}

	srcRoot, err := utils.DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		t.Error(err)
	}

	// Test unexecutable plugin
	unexecutablePlugin := filepath.Join(
		srcRoot, "builtin", "", "secretgenerator", "SecretGenerator.so")
	p := NewExecPlugin(unexecutablePlugin)
	err = p.ErrIfNotExecutable()
	if err == nil {
		t.Fatalf("unexpected err: %v", err)
	}

	// Test executable plugin
	executablePlugin := filepath.Join(
		srcRoot, "someteam.example.com", "v1", "bashedconfigmap", "BashedConfigMap")
	p = NewExecPlugin(executablePlugin)
	err = p.ErrIfNotExecutable()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
