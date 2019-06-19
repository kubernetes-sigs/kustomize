// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kunstruct

import (
	"reflect"
	"testing"
)

type PathSectionSlice []PathSection

func buildPath(idx int, fields ...string) PathSectionSlice {
	return PathSectionSlice{PathSection{fields: fields, idx: idx}}
}

func (a PathSectionSlice) addSearchKey(name string, value string) PathSectionSlice {
	a[len(a)-1].searchName = name
	a[len(a)-1].searchValue = value
	return a
}

func (a PathSectionSlice) addSection(idx int, fields ...string) PathSectionSlice {
	return append(a, PathSection{fields: fields, idx: idx})
}

func TestParseField(t *testing.T) {
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
			expectedValue: buildPath(-1, "Kind"),
			errorExpected: false,
		},
		{
			name:          "twoFields",
			pathToField:   "metadata.name",
			expectedValue: buildPath(-1, "metadata", "name"),
			errorExpected: false,
		},
		{
			name:          "threeFields",
			pathToField:   "spec.ports.port",
			expectedValue: buildPath(-1, "spec", "ports", "port"),
			errorExpected: false,
		},
		{
			name:          "empty",
			pathToField:   "",
			expectedValue: buildPath(-1, ""),
			errorExpected: false,
		},
		{
			name:          "validStringIndex",
			pathToField:   "that[1]",
			expectedValue: buildPath(1, "that"),
			errorExpected: false,
		},
		{
			name:          "sliceInSlice",
			pathToField:   "that[1][0]",
			expectedValue: buildPath(1, "that").addSection(0),
			errorExpected: false,
		},
		{
			name:          "validStructSubField",
			pathToField:   "those[1].field2",
			expectedValue: buildPath(1, "those").addSection(-1, "field2"),
			errorExpected: false,
		},
		{
			name:          "validStructSubFieldIndex",
			pathToField:   "these[1].field2[0]",
			expectedValue: buildPath(1, "these").addSection(0, "field2"),
			errorExpected: false,
		},
		{
			name:          "validStructSubFieldIndexSubfield",
			pathToField:   "complextree[1].field2[1].stringsubfield",
			expectedValue: buildPath(1, "complextree").addSection(1, "field2").addSection(-1, "stringsubfield"),
			errorExpected: false,
		},
		{
			name:          "validStructDownwardAPI",
			pathToField:   `metadata.labels["app.kubernetes.io/component"]`,
			expectedValue: buildPath(-1, "metadata", "labels", "app.kubernetes.io/component"),
			errorExpected: false,
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
			expectedValue: buildPath(1, "complextree").addSection(1, "field2").addSection(-1, "stringsubfield"),
			errorExpected: false,
		},
		{
			name:        "validStructDownwardAPI2",
			pathToField: `spec.template.spec.containers[name=main].image`,
			expectedValue: buildPath(-1, "spec", "template", "spec", "containers").
				addSearchKey("name", "main").
				addSection(-1, "image"),
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
