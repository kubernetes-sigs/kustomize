package krusty_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestComposition_Basics(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("cm.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    happy: "true"
  name: my-config
data:
  MY_ENV: foo
`)
	th.WriteComposition(".", `
transformers:
- apiVersion: builtin
  kind: ResourceGenerator
  files:
  - cm.yaml
- apiVersion: builtin
  kind: PrefixSuffixTransformer
  metadata:
    name: myFancyNamePrefixer
  prefix: bob-
  fieldSpecs:
  - path: metadata/name
  - path: data/MY_ENV
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
data:
  MY_ENV: bob-foo
kind: ConfigMap
metadata:
  annotations:
    happy: "true"
  name: bob-my-config
`)
}

func TestComposition_SuccessDir(t *testing.T) {
	expectedTestCases := 16
	testCasesFound := 0
	testCaseDir := filepath.FromSlash("testdata/composition_testcases/success")

	err := filepath.Walk(testCaseDir, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() {
			return nil
		}

		// If the expected.yaml file doesn't exist, we go into this directory's children.
		goldenFile := filepath.Join(path, "expected.yaml")
		if _, err := os.Stat(goldenFile); os.IsNotExist(err) {
			return nil
		}
		require.NoError(t, err, "could not stat expected results file")

		t.Run(path, func(t *testing.T) {
			expected, err := ioutil.ReadFile(goldenFile)
			require.NoError(t, err, "failed to load expected results file")

			compositionFile := filepath.Join(path, "composition.yaml")
			_, err = os.Stat(compositionFile)
			require.NoError(t, err, "could not stat composition file")

			testCasesFound += 1
			th := kusttest_test.MakeHarness(t)
			err = th.CopyAllToFs(filepath.Dir(path))
			require.NoError(t, err, "failed copying to test FS")

			opts := th.MakeOptionsPluginsEnabled()
			// TODO: Should this not be using statically linked builtin plugins?
			opts.PluginConfig.BpLoadingOptions = types.BploUseStaticallyLinked

			// TODO: add assertions (somewhere, maybe here) showing flattened Compositions

			m := th.Run(filepath.Dir(compositionFile), opts)
			th.AssertActualEqualsExpected(
				m, string(expected))
		})

		// If the expected.yaml existed, we skip all children directories.
		return filepath.SkipDir
	})
	assert.NoError(t, err, "Failed walking test directories")
	require.Equal(t, expectedTestCases, testCasesFound, "unexpected number of test cases")
}

func TestComposition_ErrorDir(t *testing.T) {
	expectedTestCases := 14
	testCasesFound := 0
	testCaseDir := filepath.FromSlash("testdata/composition_testcases/error")

	err := filepath.Walk(testCaseDir, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() {
			return nil
		}

		// If the errors.txt file doesn't exist, we go into this directory's children.
		errorsFile := filepath.Join(path, "errors.txt")
		if _, err := os.Stat(errorsFile); os.IsNotExist(err) {
			return nil
		}
		require.NoError(t, err, "could not stat expected errors file")

		t.Run(path, func(t *testing.T) {
			expected, err := ioutil.ReadFile(errorsFile)
			require.NoError(t, err, "failed to load expected errors file")

			compositionFile := filepath.Join(path, "composition.yaml")
			_, err = os.Stat(compositionFile)
			require.NoError(t, err, "could not stat composition file")

			testCasesFound += 1
			th := kusttest_test.MakeHarness(t)
			err = th.CopyAllToFs(filepath.Dir(path))
			require.NoError(t, err, "failed copying to test FS")

			opts := th.MakeOptionsPluginsEnabled()
			// TODO: Should this not be using statically linked builtin plugins?
			opts.PluginConfig.BpLoadingOptions = types.BploUseStaticallyLinked

			actualErr := th.RunWithErr(filepath.Dir(compositionFile), opts)
			for _, msg := range strings.Split(string(expected), "\n") {
				require.Regexp(t, regexp.MustCompile(msg), actualErr.Error())
			}
		})

		// If the expected.yaml existed, we skip all children directories.
		return filepath.SkipDir
	})
	assert.NoError(t, err, "Failed walking test directories")
	require.Equal(t, expectedTestCases, testCasesFound, "unexpected number of test cases")
}
