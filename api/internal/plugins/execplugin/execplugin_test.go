// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	fLdr "sigs.k8s.io/kustomize/api/internal/loader"
	. "sigs.k8s.io/kustomize/api/internal/plugins/execplugin"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/plugins/utils"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	expectedLargeConfigMap = `{"apiVersion":"v1","data":{"password":"password","username":"user"},"kind":"ConfigMap",` +
		`"metadata":{"annotations":{"internal.config.kubernetes.io/generatorBehavior":"unspecified",` +
		`"internal.config.kubernetes.io/needsHashSuffix":"enabled"},"name":"example-configmap-test"}}`
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
		fLdr.RestrictionRootOnly, filesys.Separator, fSys)
	if err != nil {
		t.Fatal(err)
	}
	pvd := provider.NewDefaultDepProvider()
	rf := resmap.NewFactory(pvd.GetResourceFactory())
	pluginConfig, err := rf.RF().FromMap(
		map[string]interface{}{
			"apiVersion": "someteam.example.com/v1",
			"kind":       "SedTransformer",
			"metadata": map[string]interface{}{
				"name": "some-random-name",
			},
			"argsOneLiner": "one two 'foo bar'",
			"argsFromFile": "sed-input.txt",
		})
	if err != nil {
		t.Fatalf("failed to writes the data to a file: %v", err)
	}

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

// TestExecPluginLarge loads PluginConfigs of various (large) sizes. It tests if the env variable is kept below the
// maximum of 131072 bytes.
func TestExecPluginLarge(t *testing.T) {
	// Skip this test on windows.
	if runtime.GOOS == "windows" {
		t.Skipf("always returns nil on Windows")
	}

	// Add executable plugin.
	srcRoot, err := utils.DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		t.Error(err)
	}
	executablePlugin := filepath.Join(
		srcRoot, "someteam.example.com", "v1", "bashedconfigmap", "BashedConfigMap")
	p := NewExecPlugin(executablePlugin)
	err = p.ErrIfNotExecutable()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	// Create a fake filesystem.
	fSys := filesys.MakeFsInMemory()

	// Load plugin config.
	ldr, err := fLdr.NewLoader(
		fLdr.RestrictionRootOnly, filesys.Separator, fSys)
	if err != nil {
		t.Fatal(err)
	}
	pvd := provider.NewDefaultDepProvider()
	rf := resmap.NewFactory(pvd.GetResourceFactory())
	pc := types.DisabledPluginConfig()

	// Test for various lengths. 131071 is the max length that we can have for any given env var in Bytes.
	tcs := []struct {
		length int
		char   rune
	}{
		{1000, 'a'},
		{131071, 'a'},
		{131072, 'a'},
		{200000, 'a'},
		{131071, '安'},
		{131074, '安'},
	}
	for _, tc := range tcs {
		t.Logf("Testing with an env var length of %d and character %c", tc.length, tc.char)
		pluginConfig, err := rf.RF().FromBytes(buildLargePluginConfig(tc.length, tc.char))
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		yaml, err := pluginConfig.AsYAML()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		err = p.Config(resmap.NewPluginHelpers(ldr, pvd.GetFieldValidator(), rf, pc), yaml)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		resMap, err := p.Generate()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		rNodeSlices := resMap.ToRNodeSlice()
		for _, rNodeSlice := range rNodeSlices {
			json, err := rNodeSlice.MarshalJSON()
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if string(json) != expectedLargeConfigMap {
				t.Fatalf("expected generated JSON to match %q, but got %q instead",
					expectedLargeConfigMap, string(json))
			}
		}
	}
}

// buildLargePluginConfig builds a plugin configuration of length: length - len("KUSTOMIZE_PLUGIN_CONFIG_STRING=")
// This allows us to create an environment variable KUSTOMIZE_PLUGIN_CONFIG_STRING=<plugin content> with the exact
// length that's provided in the length parameter. Used as a helper for TestExecPluginLarge.
func buildLargePluginConfig(length int, char rune) []byte {
	length -= len("KUSTOMIZE_PLUGIN_CONFIG_STRING=")

	var sb strings.Builder
	sb.WriteString("apiVersion: someteam.example.com/v1\n")
	sb.WriteString("kind: BashedConfigMap\n")
	sb.WriteString("metadata:\n")
	sb.WriteString("  name: some-random-name\n")
	sb.WriteString("argsOneLiner: \"user password\"\n")
	sb.WriteString("customArg: ")

	// Now, fill up parameter customArg: until we reach the desired length. Account for the fact that runes can be
	// 1 to 4 Bytes each.
	upperBound := length - sb.Len()
	for i := 0; i < upperBound-len(string(char)); i += len(string(char)) {
		sb.WriteRune(char)
	}
	return []byte(sb.String())
}
