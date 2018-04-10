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

package commands

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ghodss/yaml"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
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
		{"path", []string{"too", "many"}, "", "specify one path to manifest"},
	}
	for _, mycase := range cases {
		opts := buildOptions{}
		e := opts.Validate(mycase.args)
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
		if opts.manifestPath != mycase.path {
			t.Errorf("%s: expected path '%s', got '%s'", mycase.name, mycase.path, opts.manifestPath)
		}
	}
}

func TestBuild(t *testing.T) {
	const updateEnvVar = "UPDATE_KUSTOMIZE_EXPECTED_DATA"
	updateKustomizeExpected := os.Getenv(updateEnvVar) == "true"
	fs := fs.MakeRealFS()

	testcases := sets.NewString()
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
				testcases.Insert(strings.TrimPrefix(name, "testcase-"))
			}
			return filepath.SkipDir
		}
		return nil
	})
	// sanity check that we found the right folder
	if !testcases.Has("simple") {
		t.Fatalf("Error locating testcases")
	}

	for _, testcaseName := range testcases.List() {
		t.Run(testcaseName, func(t *testing.T) {
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
				manifestPath: testcase.Filename,
			}
			buf := bytes.NewBuffer([]byte{})
			err = ops.RunBuild(buf, os.Stderr, fs)
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
					t.Errorf("%s\ndoesn't equal expected:\n%s\n", actualBytes, expectedBytes)
				}
			} else {
				ioutil.WriteFile(testcase.ExpectedStdout, actualBytes, 0644)
			}

		})
	}

}
