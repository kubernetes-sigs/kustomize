// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func shouldContain(t *testing.T, s []byte, x string) {
	if !strings.Contains(string(s), x) {
		t.Fatalf("unable to find %s", x)
	}
}

func TestPrintPluginEnvPlugin(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "PrintPluginEnv")
	// Just to be clear.
	th.ResetLoaderRoot("/theAppRoot")
	defer th.Reset()

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: PrintPluginEnv
metadata:
  name: whatever
`)
	a, err := m.AsYaml()
	if err != nil {
		t.Error(err)
	}
	shouldContain(t, a, "kustomize_plugin_config_root: /theAppRoot")
	shouldContain(t, a, "plugin/someteam.example.com/v1/printpluginenv")
}
