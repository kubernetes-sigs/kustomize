// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func shouldContain(t *testing.T, s []byte, x string) {
	if !strings.Contains(string(s), x) {
		t.Fatalf("unable to find %s", x)
	}
}

func TestPrintWorkDirPlugin(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "PrintWorkDir")

	th := kusttest_test.NewKustTestHarnessAllowPlugins(t, "/theAppRoot")

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: PrintWorkDir
metadata:
  name: whatever
`)
	a, err := m.AsYaml()
	if err != nil {
		t.Error(err)
	}
	shouldContain(t, a, "path: /theAppRoot")
	shouldContain(t, a, "plugin/someteam.example.com/v1/printworkdir")
}
