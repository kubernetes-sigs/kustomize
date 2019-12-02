// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/internal/loadertest"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func makeKustTarget(
	t *testing.T,
	fSys filesys.FileSystem,
	root string) *target.KustTarget {
	return makeKustTargetWithRf(
		t, fSys, root,
		resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
}

func makeKustTargetWithRf(
	t *testing.T,
	fSys filesys.FileSystem,
	root string,
	resFact *resource.Factory) *target.KustTarget {
	rf := resmap.NewFactory(resFact, transformer.NewFactoryImpl())
	pc := konfig.DisabledPluginConfig()
	kt := target.NewKustTarget(
		loadertest.NewFakeLoaderWithRestrictor(
			loader.RestrictionRootOnly, fSys, root),
		valtest_test.MakeFakeValidator(),
		rf,
		transformer.NewFactoryImpl(),
		pLdr.NewLoader(pc, rf))
	err := kt.Load()
	if err != nil {
		t.Fatalf("Unexpected construction error %v", err)
	}
	return kt
}
