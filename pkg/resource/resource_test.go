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
)

func TestGetFieldAsString(t *testing.T) {
	m := map[string]interface{}{
		"Kind": "Service",
		"metadata": map[string]interface{}{
			"labels": map[string]string{
				"app": "application-name",
			},
			"name": "service-name",
		},
	}

	tests := []struct {
		pathToField   []string
		expectedName  string
		expectedError bool
	}{
		{
			pathToField:   []string{"Kind"},
			expectedName:  "Service",
			expectedError: false,
		},
		{
			pathToField:   []string{"metadata", "name"},
			expectedName:  "service-name",
			expectedError: false,
		},
		{
			pathToField:   []string{"metadata", "non-existing-field"},
			expectedName:  "",
			expectedError: true,
		},
	}

	for _, test := range tests {
		s, err := GetFieldValue(m, test.pathToField)
		if test.expectedError && err == nil {
			t.Fatalf("should return error, but no error returned")
		} else {
			if test.expectedName != s {
				t.Fatalf("Got:%s expected:%s", s, test.expectedName)
			}
		}
	}
}
