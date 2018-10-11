/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package build

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"sigs.k8s.io/kustomize/internal/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/commands/kustfile"
	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/fs"
)

type buildTestCase struct {
	Description string   `yaml:"description"`
	Args        []string `yaml:"args"`
	Filename    string   `yaml:"filename"`
	// path to the file that contains the expected output
	ExpectedStdout string `yaml:"expectedStdout"`
	ExpectedError  string `yaml:"expectedError"`
}

func TestBuildValidate(t *testing.T) {
	var cases = []struct {
		name  string
		args  []string
		path  string
		erMsg string
	}{
		{"noargs", []string{}, "./", ""},
		{"file", []string{"beans"}, "beans", ""},
		{"path", []string{"a/b/c"}, "a/b/c", ""},
		{"path", []string{"too", "many"},
			"", "specify one path to " + constants.KustomizationFileName},
	}
	for _, mycase := range cases {
		opts := buildOptions{}
		e := opts.Validate(mycase.args, "", nil)
		if len(mycase.erMsg) > 0 {
			if e == nil {
				t.Errorf("%s: Expected an error %v", mycase.name, mycase.erMsg)
			}
			if e.Error() != mycase.erMsg {
				t.Errorf("%s: Expected error %s, but got %v", mycase.name, mycase.erMsg, e)
			}
			continue
		}
		if e != nil {
			t.Errorf("%s: unknown error %v", mycase.name, e)
			continue
		}
		if opts.kustomizationPath != mycase.path {
			t.Errorf("%s: expected path '%s', got '%s'", mycase.name, mycase.path, opts.kustomizationPath)
		}
	}
}

func TestBuild(t *testing.T) {
	const updateEnvVar = "UPDATE_KUSTOMIZE_EXPECTED_DATA"
	updateKustomizeExpected := os.Getenv(updateEnvVar) == "true"
	fSys := fs.MakeRealFS()

	var testcases []string
	filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == "testdata" {
			return nil
		}
		name := filepath.Base(path)
		if info.IsDir() {
			if strings.HasPrefix(name, "testcase-") {
				testcases = append(testcases, strings.TrimPrefix(name, "testcase-"))
			}
			return filepath.SkipDir
		}
		return nil
	})
	// sanity check that we found the right folder
	if !kustfile.StringInSlice("simple", testcases) {
		t.Fatalf("Error locating testcases")
	}

	for _, testcaseName := range testcases {
		t.Run(testcaseName,
			func(t *testing.T) {
				runBuildTestCase(t, testcaseName, updateKustomizeExpected, fSys)
			})
	}

}

func runBuildTestCase(t *testing.T, testcaseName string, updateKustomizeExpected bool, fSys fs.FileSystem) {
	name := testcaseName
	testcase := buildTestCase{}
	testcaseDir := filepath.Join("testdata", "testcase-"+name)
	testcaseData, err := ioutil.ReadFile(filepath.Join(testcaseDir, "test.yaml"))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if err := yaml.Unmarshal(testcaseData, &testcase); err != nil {
		t.Fatalf("%s: %v", name, err)
	}

	ops := &buildOptions{
		kustomizationPath: testcase.Filename,
	}
	buf := bytes.NewBuffer([]byte{})
	f := k8sdeps.NewFactory()
	err = ops.RunBuild(
		buf, fSys,
		f.ResmapF,
		f.TransformerF)
	switch {
	case err != nil && len(testcase.ExpectedError) == 0:
		t.Errorf("unexpected error: %v", err)
	case err != nil && len(testcase.ExpectedError) != 0:
		if !strings.Contains(err.Error(), testcase.ExpectedError) {
			t.Errorf("expected error to contain %q but got: %v", testcase.ExpectedError, err)
		}
		return
	case err == nil && len(testcase.ExpectedError) != 0:
		t.Errorf("unexpected no error")
	}

	actualBytes := buf.Bytes()
	if !updateKustomizeExpected {
		expectedBytes, err := ioutil.ReadFile(testcase.ExpectedStdout)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(actualBytes, expectedBytes) {
			t.Errorf("\n**** Actual:\n\n%s\n\n**** doesn't equal expected:\n\n%s\n\n", actualBytes, expectedBytes)
		}
	} else {
		ioutil.WriteFile(testcase.ExpectedStdout, actualBytes, 0644)
	}

}
