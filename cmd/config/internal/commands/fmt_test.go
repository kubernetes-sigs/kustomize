// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/kio/filters/testyaml"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/testutil"
)

// TestCmd_files verifies the fmt command formats the files
func TestFmtCommand_files(t *testing.T) {
	f1, err := ioutil.TempFile("", "cmdfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(f1.Name())
	err = ioutil.WriteFile(f1.Name(), testyaml.UnformattedYaml1, 0600)
	if !assert.NoError(t, err) {
		return
	}

	f2, err := ioutil.TempFile("", "cmdfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(f2.Name())
	err = ioutil.WriteFile(f2.Name(), testyaml.UnformattedYaml2, 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	r := commands.GetFmtRunner("")
	r.Command.SetArgs([]string{f1.Name(), f2.Name()})
	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	// verify the contents
	b, err := ioutil.ReadFile(f1.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, string(testyaml.FormattedYaml1), string(b)) {
		return
	}

	b, err = ioutil.ReadFile(f2.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, string(testyaml.FormattedYaml2), string(b)) {
		return
	}
}

func TestFmtCommand_stdin(t *testing.T) {
	out := &bytes.Buffer{}
	r := commands.GetFmtRunner("")
	r.Command.SetOut(out)
	r.Command.SetIn(bytes.NewReader(testyaml.UnformattedYaml1))

	// fmt the input
	err := r.Command.Execute()
	assert.NoError(t, err)

	// verify the output
	assert.Contains(t, out.String(), string(testyaml.FormattedYaml1))
}

// TestCmd_filesAndstdin verifies that if both files and stdin input are provided, only
// the files are formatted and the input is ignored
func TestFmtCmd_filesAndStdin(t *testing.T) {
	f1, err := ioutil.TempFile("", "cmdfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	err = ioutil.WriteFile(f1.Name(), testyaml.UnformattedYaml1, 0600)
	if !assert.NoError(t, err) {
		return
	}

	f2, err := ioutil.TempFile("", "cmdfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	err = ioutil.WriteFile(f2.Name(), testyaml.UnformattedYaml2, 0600)
	if !assert.NoError(t, err) {
		return
	}

	out := &bytes.Buffer{}
	in := &bytes.Buffer{}
	r := commands.GetFmtRunner("")
	r.Command.SetOut(out)
	r.Command.SetIn(in)

	// fmt the files
	r = commands.GetFmtRunner("")
	r.Command.SetArgs([]string{f1.Name(), f2.Name()})
	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	// verify the output
	b, err := ioutil.ReadFile(f1.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, string(testyaml.FormattedYaml1), string(b)) {
		return
	}

	b, err = ioutil.ReadFile(f2.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, string(testyaml.FormattedYaml2), string(b)) {
		return
	}
	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Equal(t, "", out.String()) {
		return
	}
}

// TestCmd_files verifies the fmt command formats the files
func TestCmd_failFiles(t *testing.T) {
	// fmt the files
	r := commands.GetFmtRunner("")
	r.Command.SetArgs([]string{"notrealfile"})
	r.Command.SilenceUsage = true
	r.Command.SilenceErrors = true
	err := r.Command.Execute()
	testutil.AssertErrorContains(t, err, "notrealfile:")
}

// TestCmd_files verifies the fmt command formats the files
func TestCmd_failFileContents(t *testing.T) {
	out := &bytes.Buffer{}
	r := commands.GetFmtRunner("")
	r.Command.SetOut(out)
	r.Command.SetIn(strings.NewReader(`{`))

	// fmt the input
	err := r.Command.Execute()

	// expect an error
	assert.EqualError(t, err, "MalformedYAMLError: yaml: line 1: did not find expected node content")
}

func TestFmtSubPackages(t *testing.T) {
	var tests = []struct {
		name        string
		dataset     string
		packagePath string
		args        []string
		expected    string
	}{
		{
			name:    "fmt-recurse-subpackages",
			dataset: "dataset-with-setters",
			args:    []string{"-R"},
			expected: `${baseDir}/
formatted resource files in the package

${baseDir}/mysql/
formatted resource files in the package

${baseDir}/mysql/nosetters/
formatted resource files in the package

${baseDir}/mysql/storage/
formatted resource files in the package
`,
		},
		{
			name:        "fmt-top-level-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			packagePath: "mysql",
			expected: `${baseDir}/mysql/
formatted resource files in the package
`,
		},
		{
			name:        "fmt-nested-pkg-no-recurse-subpackages",
			dataset:     "dataset-without-setters",
			packagePath: "mysql/storage",
			expected: `${baseDir}/mysql/storage/
formatted resource files in the package
`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// reset the openAPI afterward
			openapi.ResetOpenAPI()
			defer openapi.ResetOpenAPI()
			sourceDir := filepath.Join("test", "testdata", test.dataset)
			baseDir := t.TempDir()
			copyutil.CopyDir(sourceDir, baseDir)
			runner := commands.GetFmtRunner("")
			actual := &bytes.Buffer{}
			runner.Command.SetOut(actual)
			runner.Command.SetArgs(append([]string{filepath.Join(baseDir, test.packagePath)}, test.args...))
			err := runner.Command.Execute()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// normalize path format for windows
			actualNormalized := strings.ReplaceAll(
				strings.ReplaceAll(actual.String(), "\\", "/"),
				"//", "/")

			expected := strings.ReplaceAll(test.expected, "${baseDir}", baseDir)
			expectedNormalized := strings.ReplaceAll(expected, "\\", "/")
			if !assert.Contains(t, strings.TrimSpace(actualNormalized), strings.TrimSpace(expectedNormalized)) {
				t.FailNow()
			}
		})
	}
}
