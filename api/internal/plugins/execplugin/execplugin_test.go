// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	. "sigs.k8s.io/kustomize/api/internal/plugins/execplugin"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/plugins/utils"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestExecPluginConfig(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile("sed-input.txt", []byte(`
s/$FOO/foo/g
s/$BAR/bar baz/g
 \ \ \ 
`))
	ldr, err := fLdr.NewLoader(
		fLdr.RestrictionRootOnly, filesys.Separator, fSys)
	if err != nil {
		t.Fatal(err)
	}
	rf := resmap.NewFactory(
		resource.NewFactory(
			kunstruct.NewKunstructuredFactoryImpl()), nil)
	v := valtest_test.MakeFakeValidator()
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

	p := NewExecPlugin(
		pLdr.AbsolutePluginPath(
			konfig.DisabledPluginConfig(),
			pluginConfig.OrgId()))
	// Not checking to see if the plugin is executable,
	// because this test does not run it.
	// This tests only covers sending configuration
	// to the plugin wrapper object and confirming
	// that it's properly prepared for execution.

	yaml, err := pluginConfig.AsYAML()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	p.Config(resmap.NewPluginHelpers(ldr, v, rf), yaml)

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
