// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/loadertest"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/k8sdeps/transformer"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/pgmconfig"
	"sigs.k8s.io/kustomize/api/plugins/builtinconfig/consts"
	"sigs.k8s.io/kustomize/api/plugins/config"
	pLdr "sigs.k8s.io/kustomize/api/plugins/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/target"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
)

// KustTestHarness is an environment for running a kustomize build,
// aka a run of MakeCustomizedResMap.  It holds a file loader
// presumably primed with an in-memory file system, a plugin
// loader, factories to make what it needs, etc.
type KustTestHarness struct {
	t   *testing.T
	rf  *resmap.Factory
	ldr loadertest.FakeLoader
	pl  *pLdr.Loader

	// tempDir is the temporary on-disk directory used by the test.  Only set
	// if a DiskFS is used.
	tempDir string
}

func NewKustTestHarness(t *testing.T, path string) *KustTestHarness {
	return NewKustTestHarnessFull(
		t, path, fLdr.RestrictionRootOnly, config.DefaultPluginConfig())
}

func NewKustTestHarnessAllowPlugins(t *testing.T, path string) *KustTestHarness {
	return NewKustTestHarnessFull(
		t, path, fLdr.RestrictionRootOnly, config.ActivePluginConfig())
}

func NewKustTestHarnessNoLoadRestrictor(t *testing.T, path string) *KustTestHarness {
	return NewKustTestHarnessFull(
		t, path, fLdr.RestrictionNone, config.DefaultPluginConfig())
}

func NewKustTestHarnessFull(
	t *testing.T, path string,
	lr fLdr.LoadRestrictorFunc, pc *types.PluginConfig) *KustTestHarness {
	rf := resmap.NewFactory(resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	return &KustTestHarness{
		t:   t,
		rf:  rf,
		ldr: loadertest.NewFakeLoaderWithRestrictor(lr, path),
		pl:  pLdr.NewLoader(pc, rf)}
}

// NewKustTestHarnessDiskFS returns a new test harness backed by a disk FS.
func NewKustTestHarnessDiskFS(t *testing.T, path string) *KustTestHarness {
	// create the temp directory backed by a real fs -- cleaned up by calling Cleanup()
	fSys := filesys.MakeFsOnDisk()
	dir, err := ioutil.TempDir("", "kustomize-test-harness")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	path = filepath.Join(dir, path)
	if !assert.NoError(t, fSys.Mkdir(path)) {
		_ = os.RemoveAll(dir)
		t.FailNow()
	}

	// create the loader
	ldr, err := loadertest.NewFakeLoaderWithFS(fSys, fLdr.RestrictionRootOnly, path)
	if !assert.NoError(t, err) {
		_ = os.RemoveAll(dir)
		t.FailNow()
	}

	// create the resource map
	rf := resmap.NewFactory(resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

	// create the test harness
	return &KustTestHarness{
		tempDir: dir, // save the directory so it can be cleaned up at the end of the test
		t:       t,
		rf:      rf,
		ldr:     ldr,
		pl:      pLdr.NewLoader(config.DefaultPluginConfig(), rf)}
}

// FromRoot returns the path from the test harness root directory
func (th *KustTestHarness) FromRoot(path string) string {
	return filepath.Join(th.ldr.Root(), path)
}

// Cleanup cleans up any temporary files created by the test harness
func (th *KustTestHarness) Cleanup() error {
	if strings.HasPrefix(th.tempDir, os.TempDir()) {
		// clean up the temps directory if it was initialized
		return os.RemoveAll(th.tempDir)
	}
	return nil
}

func (th *KustTestHarness) MakeKustTarget() *target.KustTarget {
	kt, err := target.NewKustTarget(
		th.ldr, valtest_test.MakeFakeValidator(), th.rf,
		transformer.NewFactoryImpl(), th.pl)
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	return kt
}

func (th *KustTestHarness) WriteF(dir string, content string) {
	err := th.ldr.AddFile(dir, []byte(content))
	if err != nil {
		th.t.Fatalf("failed write to %s; %v", dir, err)
	}
}

func (th *KustTestHarness) WriteK(dir string, content string) {
	th.WriteF(
		filepath.Join(
			dir,
			pgmconfig.DefaultKustomizationFileName()), `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`+content)
}

func (th *KustTestHarness) RF() *resource.Factory {
	return th.rf.RF()
}

func (th *KustTestHarness) FromMap(m map[string]interface{}) *resource.Resource {
	return th.rf.RF().FromMap(m)
}

func (th *KustTestHarness) FromMapAndOption(m map[string]interface{}, args *types.GeneratorArgs, option *types.GeneratorOptions) *resource.Resource {
	return th.rf.RF().FromMapAndOption(m, args, option)
}

func (th *KustTestHarness) WriteDefaultConfigs(fName string) {
	m := consts.GetDefaultFieldSpecsAsMap()
	var content []byte
	for _, tCfg := range m {
		content = append(content, []byte(tCfg)...)
	}
	err := th.ldr.AddFile(fName, content)
	if err != nil {
		th.t.Fatalf("unable to add file %s", fName)
	}
}

func (th *KustTestHarness) LoadAndRunGenerator(
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

func (th *KustTestHarness) LoadAndRunTransformer(
	config, input string) resmap.ResMap {
	resMap, err := th.RunTransformer(config, input)
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	return resMap
}

func (th *KustTestHarness) ErrorFromLoadAndRunTransformer(
	config, input string) error {
	_, err := th.RunTransformer(config, input)
	return err
}

func (th *KustTestHarness) RunTransformer(
	config, input string) (resmap.ResMap, error) {
	resMap, err := th.rf.NewResMapFromBytes([]byte(input))
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	return th.RunTransformerFromResMap(config, resMap)
}

func (th *KustTestHarness) RunTransformerFromResMap(
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

func (th *KustTestHarness) AssertActualEqualsExpected(
	m resmap.ResMap, expected string) {
	th.AssertActualEqualsExpectedWithTweak(m, nil, expected)
}

func (th *KustTestHarness) AssertActualEqualsExpectedWithTweak(
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
func (th *KustTestHarness) reportDiffAndFail(actual []byte, expected string) {
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
