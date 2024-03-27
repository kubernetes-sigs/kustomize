// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/ifc"
	fLdr "sigs.k8s.io/kustomize/api/internal/loader"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func makeAndLoadKustTarget(
	t *testing.T,
	fSys filesys.FileSystem,
	root string) *target.KustTarget {
	t.Helper()
	return makeAndLoadKustTargetWithLoaderOverride(t, fSys, root, nil)
}

func makeKustTargetWithRf(
	t *testing.T,
	fSys filesys.FileSystem,
	root string,
	pvd *provider.DepProvider) *target.KustTarget {
	t.Helper()
	return makeKustTargetWithRfAndLoaderOverride(t, fSys, root, pvd, nil)
}

func makeAndLoadKustTargetWithLoaderOverride(
	t *testing.T,
	fSys filesys.FileSystem,
	root string,
	ldrWrapperFn func(ifc.Loader) ifc.Loader) *target.KustTarget {
	t.Helper()
	kt := makeKustTargetWithRfAndLoaderOverride(t, fSys, root, provider.NewDefaultDepProvider(), ldrWrapperFn)
	if err := kt.Load(); err != nil {
		t.Fatalf("Unexpected load error %v", err)
	}
	return kt
}

func makeKustTargetWithRfAndLoaderOverride(
	t *testing.T,
	fSys filesys.FileSystem,
	root string,
	pvd *provider.DepProvider,
	ldrWrapperFn func(ifc.Loader) ifc.Loader) *target.KustTarget {
	t.Helper()
	baseLoader, err := fLdr.NewLoader(fLdr.RestrictionRootOnly, root, fSys)
	if err != nil {
		t.Fatal(err)
	}
	ldr := baseLoader
	if ldrWrapperFn != nil {
		ldr = ldrWrapperFn(baseLoader)
	}
	rf := resmap.NewFactory(pvd.GetResourceFactory())
	pc := types.DisabledPluginConfig()
	return target.NewKustTarget(
		ldr,
		valtest_test.MakeFakeValidator(),
		rf,
		pLdr.NewLoader(pc, rf, fSys))
}
