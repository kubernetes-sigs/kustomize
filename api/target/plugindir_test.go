// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/k8sdeps/transformer"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/plugins/config"
	pLdr "sigs.k8s.io/kustomize/api/plugins/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/target"
	"sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestPluginDir(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "PrintWorkDir")

	base, err := os.Getwd()
	if err != nil {
		t.Fatalf("err %v", err)
	}
	dir, err := ioutil.TempDir(base, "kustomize-")
	if err != nil {
		t.Fatalf("err %v", err)
	}
	defer os.RemoveAll(dir)

	fSys := filesys.MakeFsOnDisk()
	err = fSys.WriteFile(filepath.Join(dir, "kustomization.yaml"), []byte(`
generators:
- config.yaml
`))
	if err != nil {
		t.Fatalf("err %v", err)
	}
	err = fSys.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
apiVersion: someteam.example.com/v1
kind: PrintWorkDir
metadata:
  name: some-random-name
`))
	if err != nil {
		t.Fatalf("err %v", err)
	}

	ldr, err := fLdr.NewLoader(
		fLdr.RestrictionRootOnly, dir, fSys)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	rf := resmap.NewFactory(resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl()), nil)

	pl := pLdr.NewLoader(config.ActivePluginConfig(), rf)
	tg, err := target.NewKustTarget(
		ldr, valtest_test.MakeFakeValidator(), rf, transformer.NewFactoryImpl(), pl)
	if err != nil {
		t.Fatalf("err %v", err)
	}

	m, err := tg.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	th := kusttest_test.NewKustTestHarness(t, ".")
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: WorkDir
metadata:
  name: `+dir+`
spec:
  path: `+dir+`
`)
}
