// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexMatch(t *testing.T) {
	testCases := []struct {
		name             string
		valueRegex       string
		value            string
		expectedResult   bool
		expectedErrorMsg string
	}{
		{
			name:           "empty valueRegex",
			value:          "some-value",
			expectedResult: false,
		},
		{
			name:           "match regex",
			valueRegex:     "nginx-*",
			value:          "nginx-deployment",
			expectedResult: true,
		},
		{
			name:             "invalid regex",
			valueRegex:       "*-deployment",
			value:            "nginx-deployment",
			expectedResult:   false,
			expectedErrorMsg: "error parsing regexp: missing argument to repetition operator: `*`",
		},
	}

	for i := range testCases {
		test := testCases[i]
		t.Run(test.name, func(t *testing.T) {
			sr := SearchReplace{
				ValueRegex: test.valueRegex,
			}
			res, err := sr.regexMatch(test.value)
			if test.expectedErrorMsg == "" {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			} else {
				if !assert.Equal(t, test.expectedErrorMsg, err.Error()) {
					t.FailNow()
				}
			}
			if !assert.Equal(t, test.expectedResult, res) {
				t.FailNow()
			}
		})
	}
}
