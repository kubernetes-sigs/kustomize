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
	"reflect"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

func TestInputValid(t *testing.T) {
	var testcases = []struct {
		input        string
		valid        bool
		name         string
		expectedData map[string]string
	}{
		{
			input: "otters:cute",
			valid: true,
			name:  "Adds single input",
			expectedData: map[string]string{
				"otters": "cute",
			},
		},
		{
			input: "owls:great,unicorns:magical",
			valid: true,
			name:  "Adds two items",
			expectedData: map[string]string{
				"owls":     "great",
				"unicorns": "magical",
			},
		},
		{
			input: "dogs,cats",
			valid: false,
			name:  "Does not contain colon",
		},
		{
			input: ":noKey",
			valid: false,
			name:  "Missing key",
		},
		{
			input: "noValue:",
			valid: false,
			name:  "Missing value",
		},
		{
			input: "exclamation!:point",
			valid: false,
			name:  "Non-alphanumeric input",
		},
		{
			input: "123:45",
			valid: true,
			name:  "Numeric input is allowed",
			expectedData: map[string]string{
				"123": "45",
			},
		},
		{
			input: "",
			valid: false,
			name:  "Empty input",
		},
	}
	var o addMetadataOptions
	for _, tc := range testcases {

		args := []string{tc.input}
		err := o.Validate(args, label) //use label since in Validate kindofAdd is only used for error messages

		if err != nil && tc.valid {
			t.Errorf("for test case %s, unexpected cmd error: %v", tc.name, err)
		}
		if err == nil && !tc.valid {
			t.Errorf("unexpected error: expected invalid format error for test case %v", tc.name)
		}
		if err == nil && (tc.name == "Metadata keys must be unique") {
			t.Errorf("unexpected error: for test case %s, expected already there problem", tc.name)
		}

		//o.metadata should be the same as expectedData
		if tc.valid {
			if !reflect.DeepEqual(o.metadata, tc.expectedData) {
				t.Errorf("unexpeceted error: for test case %s, unexpected data was added", tc.name)
			}
		} else {
			if len(o.metadata) != 0 {
				t.Errorf("unexpeceted error: for test case %s, expected no data to be added", tc.name)
			}
		}
	}
}

func TestAddLabelNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdAddLabel(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected an error but error is: %v", err)
	}
	if err != nil && err.Error() != "must specify label" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddLabelMultipleArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	cmd := newCmdAddLabel(fakeFS)
	args := []string{"this:input", "has:spaces"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected an error but error is: %v", err)
	}
	if err != nil && err.Error() != "labels must be comma-separated, with no spaces. See help text for example" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdAddAnnotation(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected an error but error is %v", err)
	}
	if err != nil && err.Error() != "must specify annotation" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddAnnotationMultipleArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	cmd := newCmdAddAnnotation(fakeFS)
	args := []string{"this:annotation", "has:spaces"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected an error but error is %v", err)
	}
	if err != nil && err.Error() != "annotations must be comma-separated, with no spaces. See help text for example" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
