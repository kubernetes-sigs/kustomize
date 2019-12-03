// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"

	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/internal/loadertest"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

// HarnessEnhanced manages a full plugin environment for tests.
type HarnessEnhanced struct {
	Harness
	pte *pluginTestEnv
	rf  *resmap.Factory
	ldr loadertest.FakeLoader
	pl  *pLdr.Loader
}

func MakeEnhancedHarness(t *testing.T) *HarnessEnhanced {
	pte := newPluginTestEnv(t).set()

	pc, err := konfig.EnabledPluginConfig()
	if err != nil {
		t.Fatal(err)
	}

	fSys := filesys.MakeFsInMemory()

	rf := resmap.NewFactory(
		resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()),
		transformer.NewFactoryImpl())

	result := &HarnessEnhanced{
		Harness: Harness{t: t, fSys: fSys},
		pte:     pte,
		rf:      rf,
		pl:      pLdr.NewLoader(pc, rf)}
	result.ResetLoaderRoot(filesys.Separator)

	return result
}

func (th *HarnessEnhanced) Reset() {
	th.pte.reset()
}

func (th *HarnessEnhanced) BuildGoPlugin(g, v, k string) *HarnessEnhanced {
	th.pte.buildGoPlugin(g, v, k)
	return th
}

func (th *HarnessEnhanced) PrepExecPlugin(g, v, k string) *HarnessEnhanced {
	th.pte.prepExecPlugin(g, v, k)
	return th
}

func (th *HarnessEnhanced) PrepBuiltin(k string) *HarnessEnhanced {
	th.pte.buildGoPlugin(konfig.BuiltinPluginPackage, "", k)
	return th
}

func (th *HarnessEnhanced) ResetLoaderRoot(root string) {
	th.ldr = loadertest.NewFakeLoaderWithRestrictor(
		fLdr.RestrictionRootOnly, th.fSys, root)
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
		th.t.Fatalf("Err: %v", err)
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

func (th *HarnessEnhanced) ErrorFromLoadAndRunTransformer(
	config, input string) error {
	_, err := th.RunTransformer(config, input)
	return err
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
