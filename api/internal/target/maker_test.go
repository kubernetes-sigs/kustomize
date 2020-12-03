// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func makeAndLoadKustTarget(
	t *testing.T,
	fSys filesys.FileSystem,
	root string) *target.KustTarget {
	kt := makeKustTargetWithRf(t, fSys, root, provider.NewDefaultDepProvider())
	if err := kt.Load(); err != nil {
		t.Fatalf("Unexpected load error %v", err)
	}
	return kt
}

func makeKustTargetWithRf(
	t *testing.T,
	fSys filesys.FileSystem,
	root string,
	pvd *provider.DepProvider) *target.KustTarget {
	ldr, err := fLdr.NewLoader(fLdr.RestrictionRootOnly, root, fSys)
	if err != nil {
		t.Fatal(err)
	}
	rf := resmap.NewFactory(
		pvd.GetResourceFactory(), pvd.GetConflictDetectorFactory())
	pc := konfig.DisabledPluginConfig()
	return target.NewKustTarget(
		ldr,
		valtest_test.MakeFakeValidator(),
		rf,
		pLdr.NewLoader(pc, rf))
}
