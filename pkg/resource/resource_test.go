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

package resource

import (
	"testing"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewResourceFromUnstructDefaultsToRenamingUnspecified(t *testing.T) {
	var res = NewResourceFromUnstruct(unstructured.Unstructured{})

	if res.RenamingBehavior() != RenamingBehaviorUnspecified {
		t.Fatalf("Got:%s expected:%s", res.RenamingBehavior(), RenamingBehaviorUnspecified)
	}
}

func TestNewResourceWithBehaviorHasCorrectRenamingBehavior(t *testing.T) {
	tests := []RenamingBehavior {
		RenamingBehaviorHash,
		RenamingBehaviorNone,
		RenamingBehaviorUnspecified,
	}

	for _, test := range tests {
		var res, _= NewResourceWithBehavior(&unstructured.Unstructured{}, BehaviorUnspecified, test)

		if res.RenamingBehavior() != test {
			t.Fatalf("Got:%s expected:%s", res.RenamingBehavior(), test)
		}
	}
}

func TestGetFieldValue(t *testing.T) {
	res := NewResourceFromMap(map[string]interface{}{
		"Kind": "Service",
		"metadata": map[string]interface{}{
			"labels": map[string]string{
				"app": "application-name",
			},
			"name": "service-name",
		},
		"spec": map[string]interface{}{
			"ports": map[string]interface{}{
				"port": "80",
			},
		},
	})

	tests := []struct {
		pathToField   string
		expectedValue string
		errorExpected bool
	}{
		{
			pathToField:   "Kind",
			expectedValue: "Service",
			errorExpected: false,
		},
		{
			pathToField:   "metadata.name",
			expectedValue: "service-name",
			errorExpected: false,
		},
		{
			pathToField:   "metadata.non-existing-field",
			expectedValue: "",
			errorExpected: true,
		},
		{
			pathToField:   "spec.ports.port",
			expectedValue: "80",
			errorExpected: false,
		},
	}

	for _, test := range tests {
		s, err := res.GetFieldValue(test.pathToField)
		if test.errorExpected && err == nil {
			t.Fatalf("should return error, but no error returned")
		} else {
			if test.expectedValue != s {
				t.Fatalf("Got:%s expected:%s", s, test.expectedValue)
			}
		}
	}
}
