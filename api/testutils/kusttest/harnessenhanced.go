// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/ifc"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/konfig"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// HarnessEnhanced manages a full plugin environment for tests.
type HarnessEnhanced struct {
	// An instance of *testing.T, and a filesystem (likely in-memory)
	// for loading test data - plugin config, resources to transform, etc.
	Harness

	// plugintestEnv holds the plugin compiler and data needed to
	// create compilation sub-processes.
	pte *pluginTestEnv

	// rf creates Resources from byte streams.
	rf *resmap.Factory

	// A file loader using the Harness.fSys to read test data.
	ldr ifc.Loader

	// If true, wipe the ifc.loader root (not the plugin loader root)
	// as part of cleanup.
	shouldWipeLdrRoot bool

	// A plugin loader that loads plugins from a (real) file system.
	pl *pLdr.Loader
}

func MakeEnhancedHarness(t *testing.T) *HarnessEnhanced {
	r := makeBaseEnhancedHarness(t)
	r.Harness = MakeHarnessWithFs(t, filesys.MakeFsInMemory())
	// Point the Harness's file loader to the root ('/')
	// of the in-memory file system.
	r.ResetLoaderRoot(filesys.Separator)
	return r
}

func MakeEnhancedHarnessWithTmpRoot(t *testing.T) *HarnessEnhanced {
	r := makeBaseEnhancedHarness(t)
	fSys := filesys.MakeFsOnDisk()
	r.Harness = MakeHarnessWithFs(t, fSys)
	tmpDir, err := ioutil.TempDir("", "kust-testing-")
	if err != nil {
		panic("test harness cannot make tmp dir: " + err.Error())
	}
	r.ldr, err = fLdr.NewLoader(fLdr.RestrictionRootOnly, tmpDir, fSys)
	if err != nil {
		panic("test harness cannot make ldr at tmp dir: " + err.Error())
	}
	r.shouldWipeLdrRoot = true
	return r
}

func makeBaseEnhancedHarness(t *testing.T) *HarnessEnhanced {
	rf := resmap.NewFactory(
		provider.NewDefaultDepProvider().GetResourceFactory())
	return &HarnessEnhanced{
		pte: newPluginTestEnv(t).set(),
		rf:  rf,
		pl: pLdr.NewLoader(
			types.EnabledPluginConfig(types.BploLoadFromFileSys),
			rf,
			// Plugin configs are always located on disk,
			// regardless of the test harness's FS
			filesys.MakeFsOnDisk())}
}

func (th *HarnessEnhanced) ErrIfNoHelm() error {
	_, err := exec.LookPath(th.GetPluginConfig().HelmConfig.Command)
	return err
}

func (th *HarnessEnhanced) GetRoot() string {
	return th.ldr.Root()
}

func (th *HarnessEnhanced) MkDir(path string) string {
	dir := filepath.Join(th.ldr.Root(), path)
	th.GetFSys().Mkdir(dir)
	return dir
}

func (th *HarnessEnhanced) Reset() {
	if th.shouldWipeLdrRoot {
		root, _ := filepath.EvalSymlinks(th.ldr.Root())
		tmpdir, _ := filepath.EvalSymlinks(os.TempDir())

		if !strings.HasPrefix(root, tmpdir) {
			// sanity check.
			panic("something strange about th.ldr.Root() = " + th.ldr.Root())
		}
		os.RemoveAll(th.ldr.Root())
	}
	th.pte.reset()
}

func (th *HarnessEnhanced) GetPluginConfig() *types.PluginConfig {
	return th.pl.Config()
}

func (th *HarnessEnhanced) PrepBuiltin(k string) *HarnessEnhanced {
	return th.BuildGoPlugin(konfig.BuiltinPluginPackage, "", k)
}

func (th *HarnessEnhanced) BuildGoPlugin(g, v, k string) *HarnessEnhanced {
	th.pte.prepareGoPlugin(g, v, k)
	return th
}

func (th *HarnessEnhanced) PrepExecPlugin(g, v, k string) *HarnessEnhanced {
	th.pte.prepareExecPlugin(g, v, k)
	return th
}

// ResetLoaderRoot interprets its argument as an absolute directory path.
// It creates the directory, and creates the harness's file loader
// rooted in that directory.
func (th *HarnessEnhanced) ResetLoaderRoot(root string) {
	if err := th.fSys.Mkdir(root); err != nil {
		th.t.Fatal(err)
	}
	ldr, err := fLdr.NewLoader(
		fLdr.RestrictionRootOnly, root, th.fSys)
	if err != nil {
		th.t.Fatalf("Unable to make loader: %v", err)
	}
	th.ldr = ldr
}

func (th *HarnessEnhanced) LoadAndRunGenerator(
	config string) resmap.ResMap {
	res, err := th.rf.RF().FromBytes([]byte(config))
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	g, err := th.pl.LoadGenerator(
		th.ldr, valtest_test.MakeFakeValidator(), res)
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	rm, err := g.Generate()
	if err != nil {
		th.t.Fatalf("generate err: %v", err)
	}
	rm.RemoveBuildAnnotations()
	return rm
}

func (th *HarnessEnhanced) LoadAndRunGeneratorWithBuildAnnotations(
	config string) resmap.ResMap {
	res, err := th.rf.RF().FromBytes([]byte(config))
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	g, err := th.pl.LoadGenerator(
		th.ldr, valtest_test.MakeFakeValidator(), res)
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	rm, err := g.Generate()
	if err != nil {
		th.t.Fatalf("generate err: %v", err)
	}
	return rm
}

func (th *HarnessEnhanced) LoadAndRunTransformer(
	config, input string) resmap.ResMap {
	resMap, err := th.RunTransformer(config, input)
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	return resMap
}

func (th *HarnessEnhanced) RunTransformerAndCheckResult(
	config, input, expected string) {
	resMap := th.LoadAndRunTransformer(config, input)
	th.AssertActualEqualsExpectedNoIdAnnotations(resMap, expected)
}

func (th *HarnessEnhanced) ErrorFromLoadAndRunTransformer(
	config, input string) error {
	_, err := th.RunTransformer(config, input)
	return err
}

type AssertFunc func(t *testing.T, err error)

func (th *HarnessEnhanced) RunTransformerAndCheckError(
	config, input string, assertFn AssertFunc) {
	_, err := th.RunTransformer(config, input)
	assertFn(th.t, err)
}

func (th *HarnessEnhanced) RunTransformer(
	config, input string) (resmap.ResMap, error) {
	resMap, err := th.rf.NewResMapFromBytes([]byte(input))
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	return th.RunTransformerFromResMap(config, resMap)
}

func (th *HarnessEnhanced) RunTransformerFromResMap(
	config string, resMap resmap.ResMap) (resmap.ResMap, error) {
	transConfig, err := th.rf.RF().FromBytes([]byte(config))
	if err != nil {
		th.t.Logf("config: '%s'", config)
		th.t.Fatalf("Err: %v", err)
	}
	g, err := th.pl.LoadTransformer(
		th.ldr, valtest_test.MakeFakeValidator(), transConfig)
	if err != nil {
		return nil, err
	}
	err = g.Transform(resMap)
	return resMap, err
}
