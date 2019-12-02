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
// TODO: get rid of this.  Combine Harness and PluginTestEnv.
type HarnessEnhanced struct {
	Harness
	rf  *resmap.Factory
	ldr loadertest.FakeLoader
	pl  *pLdr.Loader
}

func MakeHarnessEnhanced(
	t *testing.T, path string) *HarnessEnhanced {
	pc, err := konfig.EnabledPluginConfig()
	if err != nil {
		t.Fatal(err)
	}
	fSys := filesys.MakeFsInMemory()
	rf := resmap.NewFactory(
		resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()),
		transformer.NewFactoryImpl())
	return &HarnessEnhanced{
		Harness: Harness{t: t, fSys: fSys},
		rf:      rf,
		ldr: loadertest.NewFakeLoaderWithRestrictor(
			fLdr.RestrictionRootOnly, fSys, path),
		pl: pLdr.NewLoader(pc, rf)}
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
