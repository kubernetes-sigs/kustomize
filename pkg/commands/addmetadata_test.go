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
	"testing"

	"strings"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

func TestLabelsValid(t *testing.T) {
	var testcases = []struct {
		input string
		valid bool
		name  string
	}{
		{
			input: "owls:great,unicorns:magical",
			valid: true,
			name:  "Adds two labels",
		},
		{
			input: "owls:great",
			valid: false,
			name:  "Label keys must be unique",
		},
		{
			input: "otters:cute",
			valid: true,
			name:  "Adds single label",
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
	}
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	cmd := newCmdAddLabel(fakeFS)

	for _, tc := range testcases {
		labels := strings.Split(tc.input, ",")
		//run command with test input
		args := []string{tc.input}
		err := cmd.RunE(cmd, args)

		if err != nil && tc.valid {
			t.Errorf("for test case %s, unexpected cmd error: %v", tc.name, err)
		}
		if err == nil && !tc.valid {
			t.Errorf("unexpected error: expected invalid annotation format error for test case %v", tc.name)
		}
		if err == nil && (tc.name == "Label keys must be unique") {
			t.Errorf("unexpected error: for test case %s, expected already there problem", tc.name)
		}

		content, readErr := fakeFS.ReadFile(constants.KustomizationFileName)
		if readErr != nil {
			t.Errorf("unexpected read error: %v", readErr)
		}
		//check if valid input was added to commonLabels
		for _, label := range labels {
			key := strings.Split(label, ":")[0]
			if !strings.Contains(string(content), key) && tc.valid {
				t.Errorf("unexpected error: for test case %s, expected key to be in file.", tc.name)
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

func TestAnnotationsValid(t *testing.T) {
	var testcases = []struct {
		input string
		valid bool
		name  string
	}{
		{
			input: "cats:great,dogs:okay",
			valid: true,
			name:  "Adds two annotations",
		},
		{
			input: "cats:great",
			valid: false,
			name:  "Annotation keys must be unique",
		},
		{
			input: "owls:best",
			valid: true,
			name:  "Adds single annotation",
		},
		{
			input: "cake,cookies",
			valid: false,
			name:  "Does not contain colon",
		},
		{
			input: ":hasNoKey",
			valid: false,
			name:  "Missing key",
		},
		{
			input: "hasNoValue:",
			valid: false,
			name:  "Missing value",
		},
	}
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.KustomizationFileName, []byte(kustomizationContent))
	cmd := newCmdAddAnnotation(fakeFS)

	for _, tc := range testcases {
		annotations := strings.Split(tc.input, ",")
		//run command with test input
		args := []string{tc.input}
		err := cmd.RunE(cmd, args)

		if err != nil && tc.valid {
			t.Errorf("for test case %s, unexpected cmd error: %v", tc.name, err)
		}
		if err == nil && !tc.valid {
			t.Errorf("unexpected error: expected invalid annotation format error for test case %v", tc.name)
		}
		if err == nil && (tc.name == "Annotation keys must be unique") {
			t.Errorf("unexpected error: for test case %s, expected already there problem", tc.name)
		}

		content, readErr := fakeFS.ReadFile(constants.KustomizationFileName)
		if readErr != nil {
			t.Errorf("unexpected read error: %v", readErr)
		}
		//check if valid input was added to commonAnnotations
		for _, annotation := range annotations {
			key := strings.Split(annotation, ":")[0]
			if !strings.Contains(string(content), key) && tc.valid {
				t.Errorf("unexpected error: for test case %s, expected key to be in file.", tc.name)
			}
		}

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
