// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	. "sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
)

type testingHarness struct {
	t    *testing.T
	fSys filesys.FileSystem
}

func makeTestHarness(t *testing.T) testingHarness {
	return testingHarness{
		t:    t,
		fSys: filesys.MakeFsInMemory(),
	}
}

func (th testingHarness) WriteK(path string, content string) {
	th.fSys.WriteFile(
		filepath.Join(
			path,
			konfig.DefaultKustomizationFileName()), []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`+content))
}

func (th testingHarness) WriteF(path string, content string) {
	th.fSys.WriteFile(path, []byte(content))
}

func (th testingHarness) MakeDefaultOptions() Options {
	return th.MakeOptionsPluginsDisabled()
}

func (th testingHarness) MakeOptionsPluginsDisabled() Options {
	return Options{
		LoadRestrictions: types.LoadRestrictionsRootOnly,
		PluginConfig:     konfig.DisabledPluginConfig(),
	}
}

func (th testingHarness) MakeOptionsPluginsEnabled() Options {
	c, err := konfig.EnabledPluginConfig()
	if err != nil {
		th.t.Fatal(err)
	}
	return Options{
		LoadRestrictions: types.LoadRestrictionsRootOnly,
		PluginConfig:     c,
	}
}

// Run, failing on error.
func (th testingHarness) Run(path string, o Options) resmap.ResMap {
	m, err := MakeKustomizer(th.fSys, &o).Run(path)
	if err != nil {
		th.t.Fatal(err)
	}
	return m
}

// Run, failing if there is no error.
func (th testingHarness) RunWithErr(path string, o Options) error {
	_, err := MakeKustomizer(th.fSys, &o).Run(path)
	if err == nil {
		th.t.Fatalf("expected error")
	}
	return err
}

func (th testingHarness) AssertActualEqualsExpected(
	m resmap.ResMap, expected string) {
	th.AssertActualEqualsExpectedWithTweak(m, nil, expected)
}

func (th testingHarness) AssertActualEqualsExpectedWithTweak(
	m resmap.ResMap, tweaker func([]byte) []byte, expected string) {
	if m == nil {
		th.t.Fatalf("Map should not be nil.")
	}
	// Ignore leading linefeed in expected value
	// to ease readability of tests.
	if len(expected) > 0 && expected[0] == 10 {
		expected = expected[1:]
	}
	actual, err := m.AsYaml()
	if err != nil {
		th.t.Fatalf("Unexpected err: %v", err)
	}
	if tweaker != nil {
		actual = tweaker(actual)
	}
	if string(actual) != expected {
		th.reportDiffAndFail(actual, expected)
	}
}

// Pretty printing of file differences.
func (th testingHarness) reportDiffAndFail(actual []byte, expected string) {
	sE, maxLen := convertToArray(expected)
	sA, _ := convertToArray(string(actual))
	fmt.Println("===== ACTUAL BEGIN ========================================")
	fmt.Print(string(actual))
	fmt.Println("===== ACTUAL END ==========================================")
	format := fmt.Sprintf("%%s  %%-%ds %%s\n", maxLen+4)
	limit := 0
	if len(sE) < len(sA) {
		limit = len(sE)
	} else {
		limit = len(sA)
	}
	fmt.Printf(format, " ", "EXPECTED", "ACTUAL")
	fmt.Printf(format, " ", "--------", "------")
	for i := 0; i < limit; i++ {
		fmt.Printf(format, hint(sE[i], sA[i]), sE[i], sA[i])
	}
	if len(sE) < len(sA) {
		for i := len(sE); i < len(sA); i++ {
			fmt.Printf(format, "X", "", sA[i])
		}
	} else {
		for i := len(sA); i < len(sE); i++ {
			fmt.Printf(format, "X", sE[i], "")
		}
	}
	th.t.Fatalf("Expected not equal to actual")
}

func convertToArray(x string) ([]string, int) {
	a := strings.Split(strings.TrimSuffix(x, "\n"), "\n")
	maxLen := 0
	for i, v := range a {
		z := tabToSpace(v)
		if len(z) > maxLen {
			maxLen = len(z)
		}
		a[i] = z
	}
	return a, maxLen
}

func hint(a, b string) string {
	if a == b {
		return " "
	}
	return "X"
}

func tabToSpace(input string) string {
	var result []string
	for _, i := range input {
		if i == 9 {
			result = append(result, "  ")
		} else {
			result = append(result, string(i))
		}
	}
	return strings.Join(result, "")
}
