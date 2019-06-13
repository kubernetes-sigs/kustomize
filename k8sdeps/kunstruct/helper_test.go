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

package kunstruct

import (
	"reflect"
	"testing"
)

type PathSectionSlice []PathSection

func buildPath(idx *int, fields ...string) PathSectionSlice {
	return PathSectionSlice{PathSection{fields: fields, idx: idx}}
}

func (a PathSectionSlice) addSection(idx *int, fields ...string) PathSectionSlice {
	return append(a, PathSection{fields: fields, idx: idx})
}

func TestParseField(t *testing.T) {
	i := 1
	var one *int = &i
	j := 0
	var zero *int = &j

	tests := []struct {
		name          string
		pathToField   string
		expectedValue []PathSection
		errorExpected bool
		errorMsg      string
	}{
		{
			name:          "oneField",
			pathToField:   "Kind",
			expectedValue: buildPath(nil, "Kind"),
			errorExpected: false,
		},
		{
			name:          "twoFields",
			pathToField:   "metadata.name",
			expectedValue: buildPath(nil, "metadata", "name"),
			errorExpected: false,
		},
		{
			name:          "threeFields",
			pathToField:   "spec.ports.port",
			expectedValue: buildPath(nil, "spec", "ports", "port"),
			errorExpected: false,
		},
		{
			name:          "empty",
			pathToField:   "",
			expectedValue: buildPath(nil, ""),
			errorExpected: false,
		},
		{
			name:          "validStringIndex",
			pathToField:   "that[1]",
			expectedValue: buildPath(one, "that"),
			errorExpected: false,
		},
		{
			name:          "sliceInSlice",
			pathToField:   "that[1][0]",
			expectedValue: buildPath(one, "that").addSection(zero),
			errorExpected: false,
		},
		{
			name:          "validStructSubField",
			pathToField:   "those[1].field2",
			expectedValue: buildPath(one, "those").addSection(nil, "field2"),
			errorExpected: false,
		},
		{
			name:          "validStructSubFieldIndex",
			pathToField:   "these[1].field2[0]",
			expectedValue: buildPath(one, "these").addSection(zero, "field2"),
			errorExpected: false,
		},
		{
			name:          "validStructSubFieldIndexSubfield",
			pathToField:   "complextree[1].field2[1].stringsubfield",
			expectedValue: buildPath(one, "complextree").addSection(one, "field2").addSection(nil, "stringsubfield"),
			errorExpected: false,
		},
		{
			name:          "validStructSubFieldNoneIntIndex",
			pathToField:   "complextree[thisisnotanint]",
			errorExpected: true,
			errorMsg:      "invalid index complextree[thisisnotanint]",
		},
		{
			name:          "invalidIndexInIndex",
			pathToField:   "complextree[1[0]]",
			errorExpected: true,
			errorMsg:      "nested parentheses are not allowed: complextree[1[0]]",
		},
		{
			name:          "invalidClosingBrackets",
			pathToField:   "complextree[1]]",
			errorExpected: true,
			errorMsg:      "invalid field path complextree[1]]",
		},
		{
			name:          "validFieldsWithQuotes",
			pathToField:   "'complextree'[1].field2[1].'stringsubfield'",
			expectedValue: buildPath(one, "complextree").addSection(one, "field2").addSection(nil, "stringsubfield"),
			errorExpected: false,
		},
	}

	for _, test := range tests {
		s, err := parseFields(test.pathToField)
		if test.errorExpected {
			compareExpectedParserError(t, test.name, test.pathToField, err, test.errorMsg)
			continue
		}
		if err != nil {
			unExpectedParserError(t, test.name, test.pathToField, err)
		}
		compareParserValues(t, test.name, test.pathToField, test.expectedValue, s)
	}
}

// unExpectedError function handles unexpected error
func unExpectedParserError(t *testing.T, name string, pathToField string, err error) {
	t.Fatalf("%q; path %q - unexpected error %v", name, pathToField, err)
}

// compareExpectedError compares the expectedError and the actualError return by parseFields
func compareExpectedParserError(t *testing.T, name string, pathToField string, err error, errorMsg string) {
	if err == nil {
		t.Fatalf("%q; path %q - should return error, but no error returned",
			name, pathToField)
	}

	if errorMsg != err.Error() {
		t.Fatalf("%q; path %q - expected error: \"%s\", got error: \"%v\"",
			name, pathToField, errorMsg, err.Error())
	}
}

// compareValues compares the expectedValue and actualValue returned by parseFields
func compareParserValues(t *testing.T, name string, pathToField string, expectedValue PathSectionSlice, actualValue []PathSection) {
	t.Helper()
	if len(expectedValue) != len(actualValue) {
		t.Fatalf("%q; Path: %s Got: %v Expected: %v", name, pathToField, actualValue, expectedValue)
	}

	for idx, expected := range expectedValue {
		if !reflect.DeepEqual(expected, actualValue[idx]) {
			t.Fatalf("%q; Path: %s idx: %v Fields Got: %v Expected: %v", name, pathToField, idx, actualValue[idx], expected)
		}
	}
}
