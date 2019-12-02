// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// The PrintPluginEnv plugin is a toy plugin that emits
// its working directory and some environment variables,
// to add regression protection to plugin loading logic.
func TestPluginEnvironment(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin(
			"someteam.example.com", "v1", "PrintPluginEnv")
	defer th.Reset()

	confirmBehavior(
		kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsInMemory()),
		filesys.Separator)

	dir := makeTmpDir(t)
	defer os.RemoveAll(dir)
	confirmBehavior(
		kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsOnDisk()),
		dir)
}

func confirmBehavior(th kusttest_test.Harness, dir string) {
	th.WriteK(dir, `
generators:
- config.yaml
`)
	th.WriteF(filepath.Join(dir, "config.yaml"), `
apiVersion: someteam.example.com/v1
kind: PrintPluginEnv
metadata:
  name: irrelevantHere
`)
	m := th.Run(dir, th.MakeOptionsPluginsEnabled())

	pHome, ok := os.LookupEnv(konfig.KustomizePluginHomeEnv)
	if !ok {
		th.GetT().Fatalf(
			"expected env var '%s' to be defined",
			konfig.KustomizePluginHomeEnv)
	}

	th.AssertActualEqualsExpected(m, `
apiVersion: v1
env:
  kustomize_plugin_config_root: `+dir+`
  kustomize_plugin_home: `+pHome+`
  pwd: `+dir+`
kind: GeneratedEnv
metadata:
  name: hello
`)
}

func makeTmpDir(t *testing.T) string {
	base, err := os.Getwd()
	if err != nil {
		t.Fatalf("err %v", err)
	}
	dir, err := ioutil.TempDir(base, "kustomize-tmp-test-")
	if err != nil {
		t.Fatalf("err %v", err)
	}
	return dir
}
