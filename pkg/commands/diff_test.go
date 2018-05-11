/*
Copyright 2018 The Kubernetes Authors.

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
	"regexp"
	"strings"
	"testing"

	"github.com/ghodss/yaml"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

type DiffTestCase struct {
	Description string   `yaml:"description"`
	Args        []string `yaml:"args"`
	Filename    string   `yaml:"filename"`
	// path to the file that contains the expected output
	ExpectedDiff  string `yaml:"expectedDiff"`
	ExpectedError string `yaml:"expectedError"`
}

func TestDiff(t *testing.T) {
	const updateEnvVar = "UPDATE_KUSTOMIZE_EXPECTED_DATA"
	updateKustomizeExpected := os.Getenv(updateEnvVar) == "true"

	noopDir, _ := regexp.Compile(`/tmp/noop-[0-9]*/`)
	transformedDir, _ := regexp.Compile(`/tmp/transformed-[0-9]*/`)
	timestamp, _ := regexp.Compile(`[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1]) (2[0-3]|[01][0-9]):[0-5][0-9]:[0-5][0-9].[0-9]* [+-]{1}[0-9]{4}`)

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
			testcase := DiffTestCase{}
			testcaseDir := filepath.Join("testdata", "testcase-"+name)
			testcaseData, err := ioutil.ReadFile(filepath.Join(testcaseDir, "test.yaml"))
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if err := yaml.Unmarshal(testcaseData, &testcase); err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			diffOps := &diffOptions{
				kustomizationPath: testcase.Filename,
			}
			buf := bytes.NewBuffer([]byte{})
			err = diffOps.RunDiff(buf, os.Stderr, fs)
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

			actualString := string(buf.Bytes())
			actualString = noopDir.ReplaceAllString(actualString, "/tmp/noop/")
			actualString = transformedDir.ReplaceAllString(actualString, "/tmp/transformed/")
			actualString = timestamp.ReplaceAllString(actualString, "YYYY-MM-DD HH:MM:SS")
			actualBytes := []byte(actualString)
			if !updateKustomizeExpected {
				expectedBytes, err := ioutil.ReadFile(testcase.ExpectedDiff)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(actualBytes, expectedBytes) {
					t.Errorf("%s\ndoesn't equal expected:\n%s\n", actualBytes, expectedBytes)
				}
			} else {
				ioutil.WriteFile(testcase.ExpectedDiff, actualBytes, 0644)
			}

		})
	}
}
