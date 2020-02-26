// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/setters2"
)

func TestGetValuesForMarkers(t *testing.T) {
	var tests = []struct {
		name           string
		pattern        string
		fieldValue     string
		markers        []string
		expectedError  error
		expectedOutput map[string]string
	}{
		{
			name:           "positive example",
			markers:        []string{"IMAGE", "VERSION"},
			pattern:        "something/IMAGE::VERSION/otherthing/IMAGE::VERSION/",
			fieldValue:     "something/nginx::0.1.0/otherthing/nginx::0.1.0/",
			expectedOutput: map[string]string{"IMAGE": "nginx", "VERSION": "0.1.0"},
		},
		{
			name:          "marker with different values",
			markers:       []string{"IMAGE", "VERSION"},
			pattern:       "something/IMAGE:VERSION/IMAGE",
			fieldValue:    "something/nginx:0.1.0/ubuntu",
			expectedError: errors.Errorf("marker IMAGE is found to have different values nginx and ubuntu"),
		},
		{
			name:          "unmatched pattern",
			markers:       []string{"IMAGE", "VERSION"},
			pattern:       "something/IMAGE:VERSION",
			fieldValue:    "otherthing/nginx:0.1.0",
			expectedError: errors.Errorf("unable to derive values for markers"),
		},
		{
			name:          "unmatched pattern at the end",
			markers:       []string{"IMAGE", "VERSION"},
			pattern:       "something/IMAGE:VERSION/abc",
			fieldValue:    "something/nginx:0.1.0/abcd",
			expectedError: errors.Errorf("unable to derive values for markers"),
		},
		{
			name:          "substring markers",
			markers:       []string{"IMAGE", "VERSION", "MAGE"},
			pattern:       "something/IMAGE:VERSION/abc/MAGE",
			fieldValue:    "something/nginx:0.1.0/abc/ubuntu",
			expectedError: errors.Errorf("no marker should be substring of other"),
		},
		{
			name:          "markers with no delimiters",
			markers:       []string{"IMAGE", "VERSION"},
			pattern:       "something/IMAGEVERSION/",
			fieldValue:    "something/nginx0.1.0/",
			expectedError: errors.Errorf("no delimiters between them"),
		},
		{
			name:          "unmatched delimiter",
			markers:       []string{"IMAGE", "VERSION"},
			pattern:       "something/IMAGE:^VERSION/otherthing/IMAGE::VERSION/",
			fieldValue:    "something/nginx::0.1.0/otherthing/nginx::0.1.0/",
			expectedError: errors.Errorf("unable to derive values for markers"),
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			values := []setters2.Value{}
			for _, marker := range test.markers {
				value := setters2.Value{
					Marker: marker,
				}
				values = append(values, value)
			}

			sc := SubstitutionCreator{
				Pattern:    test.pattern,
				Values:     values,
				FieldValue: test.fieldValue,
			}

			m, err := sc.GetValuesForMarkers()

			if test.expectedError == nil {
				// fail if expectedError is nil but actual error is not
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				// check if all the expected markers and values are present in actual map
				for k, v := range test.expectedOutput {
					if val, ok := m[k]; ok {
						assert.Equal(t, v, val)
					} else {
						t.FailNow()
					}
				}
			} else {
				//if expectedError is not nil, check for correctness of error message
				assert.Contains(t, err.Error(), test.expectedError.Error())
			}
		})
	}
}
