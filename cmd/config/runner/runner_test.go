// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/errors"
)

func TestExecuteCmdOnPkgs(t *testing.T) {
	var tests = []struct {
		name        string
		recurse     bool
		pkgPath     string
		needOpenAPI bool
		errMsg      string
		expectedOut string
	}{
		{
			name:        "need_Krmfile_error",
			recurse:     true,
			needOpenAPI: true,
			pkgPath:     "subpkg1/subdir1",
			errMsg:      `unable to find "Krmfile" in package`,
		},
		{
			name:        "Krmfile_not_needed_no_err",
			recurse:     true,
			needOpenAPI: false,
			pkgPath:     "subpkg1/subdir1",
			expectedOut: `${baseDir}/subpkg1/subdir1/
`,
		},
		{
			name:        "executeCmd_returns_error",
			recurse:     true,
			needOpenAPI: false,
			pkgPath:     "subpkg4",
			expectedOut: `${baseDir}/subpkg4/
`,
			errMsg: `this command returns an error if package has error.txt file`,
		},
		{
			name:        "executeCmd_prints_pkgpaths",
			recurse:     true,
			needOpenAPI: false,
			pkgPath:     "subpkg2",
			expectedOut: `${baseDir}/subpkg2/

${baseDir}/subpkg2/subpkg3/
`,
		},
	}

	dir, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	err = createTestDirStructure(dir)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			actual := &bytes.Buffer{}
			r := &TestRunner{}
			e := ExecuteCmdOnPkgs{
				NeedOpenAPI:        test.needOpenAPI,
				Writer:             actual,
				RootPkgPath:        filepath.Join(dir, test.pkgPath),
				RecurseSubPackages: test.recurse,
				CmdRunner:          r,
			}
			err := e.Execute()
			if test.errMsg == "" {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			} else {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), test.errMsg) {
					t.FailNow()
				}
			}

			// normalize path format for windows
			actualNormalized := strings.Replace(
				strings.Replace(actual.String(), "\\", "/", -1),
				"//", "/", -1)

			expected := strings.Replace(test.expectedOut, "${baseDir}", dir+"/", -1)
			expectedNormalized := strings.Replace(
				strings.Replace(expected, "\\", "/", -1),
				"//", "/", -1)
			if !assert.Equal(t, expectedNormalized, actualNormalized) {
				t.FailNow()
			}
		})
	}
}

func createTestDirStructure(dir string) error {
	/*
			Adds the folders to the input dir with following structure
			dir
			├── subpkg1
			│   ├── Krmfile
			│   └── subdir1
		  ├── subpkg4
		  │   ├── Krmfile
		  │   └── error.txt
			└── subpkg2
		      ├── subpkg3
		      │   ├── Krmfile
		      │   └── subdir2
			    └── Krmfile
	*/
	err := os.MkdirAll(filepath.Join(dir, "subpkg1/subdir1"), 0777|os.ModeDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(dir, "subpkg2/subpkg3/subdir2"), 0777|os.ModeDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(dir, "subpkg4"), 0777|os.ModeDir)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "subpkg1", "Krmfile"), []byte(""), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "subpkg2", "Krmfile"), []byte(""), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "subpkg2/subpkg3", "Krmfile"), []byte(""), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "subpkg4", "error.txt"), []byte(""), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "subpkg4", "Krmfile"), []byte(""), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "Krmfile"), []byte(""), 0777)
	if err != nil {
		return err
	}
	return nil
}

type TestRunner struct{}

func (r *TestRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	children, err := ioutil.ReadDir(pkgPath)
	if err != nil {
		return err
	}
	for _, child := range children {
		if child.Name() == "error.txt" {
			return errors.Errorf("this command returns an error if package has error.txt file")
		}
	}
	return nil
}
