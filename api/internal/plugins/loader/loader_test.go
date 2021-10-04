// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	. "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	//nolint:gosec
	secretGenerator = `
apiVersion: builtin
kind: SecretGenerator
metadata:
  name: secretGenerator
name: mySecret
behavior: merge
envFiles:
- a.env
- b.env
valueFiles:
- longsecret.txt
literals:
- FRUIT=apple
- VEGETABLE=carrot
`
	someServiceGenerator = `
apiVersion: someteam.example.com/v1
kind: SomeServiceGenerator
metadata:
  name: myServiceGenerator
service: my-service
port: "12345"
`
)

func TestLoader(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("builtin", "", "SecretGenerator").
		BuildGoPlugin("someteam.example.com", "v1", "SomeServiceGenerator")
	defer th.Reset()
	p := provider.NewDefaultDepProvider()
	rmF := resmap.NewFactory(p.GetResourceFactory())
	fsys := filesys.MakeFsInMemory()
	fLdr, err := loader.NewLoader(
		loader.RestrictionRootOnly,
		filesys.Separator, fsys)
	if err != nil {
		t.Fatal(err)
	}
	generatorConfigs, err := rmF.NewResMapFromBytes([]byte(
		someServiceGenerator + "---\n" + secretGenerator))
	if err != nil {
		t.Fatal(err)
	}
	for _, behavior := range []types.BuiltinPluginLoadingOptions{
		/* types.BploUseStaticallyLinked,
		types.BploLoadFromFileSys */} {
		c := types.EnabledPluginConfig(behavior)
		pLdr := NewLoader(c, rmF, fsys)
		if pLdr == nil {
			t.Fatal("expect non-nil loader")
		}
		_, err = pLdr.LoadGenerators(
			fLdr, valtest_test.MakeFakeValidator(), generatorConfigs)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestAbsolutePath(t *testing.T) {
	const group = "someteam.example.com"
	const version = "v1"
	const kind = "SedTransformer"

	// Set KUSTOMIZE_PLUGIN_HOME to a temporary directory and create a dummy
	// SedTransformer/SedTransformer.exe file in the expected directory.
	tempDir := t.TempDir()
	oldEnv := os.Getenv("KUSTOMIZE_PLUGIN_HOME")
	defer func() { _ = os.Setenv("KUSTOMIZE_PLUGIN_HOME", oldEnv) }()
	err := os.Setenv("KUSTOMIZE_PLUGIN_HOME", tempDir)
	if err != nil {
		t.Fatalf("Failed to set KUSTOMIZE_PLUGIN_HOME=%s: %v", tempDir, err)
	}

	execPluginDir := filepath.Join(tempDir, group, version, strings.ToLower(kind))
	err = os.MkdirAll(execPluginDir, 0775)
	if err != nil {
		t.Fatalf("failed to create the directories %s: %v", execPluginDir, err)
	}

	execPluginBinary := filepath.Join(execPluginDir, kind)
	if runtime.GOOS == "windows" {
		execPluginBinary += ".exe"
	}
	err = ioutil.WriteFile(execPluginBinary, []byte{}, 0660)
	if err != nil {
		t.Fatalf(
			"failed to write an empty file at %s: %v", execPluginBinary, err,
		)
	}

	// Create the loader for the SedTransformer exec plugin
	fSys := filesys.MakeFsOnDisk()
	pvd := provider.NewDefaultDepProvider()
	rf := resmap.NewFactory(pvd.GetResourceFactory())
	pluginConfig := rf.RF().FromMap(
		map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", group, version),
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name": "some-random-name",
			},
		},
	)
	pc := types.MakePluginConfig(
		types.PluginRestrictionsNone, types.BploUseStaticallyLinked,
	)
	loader := NewLoader(pc, rf, fSys)

	// Verify the absolute path is correct
	pluginPath, err := loader.AbsolutePluginPath(pluginConfig.OrgId())
	if err != nil {
		t.Fatalf("An error occurred when calling AbsolutePluginPath: %v", err)
	}

	if pluginPath != execPluginBinary {
		t.Fatalf(
			`Expected the plugin path of "%s", got "%s"`,
			execPluginBinary,
			pluginPath,
		)
	}
}
