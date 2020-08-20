// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kunstruct

import (
	"reflect"
	"testing"
)

var kunstructured = NewKunstructuredFactoryImpl().FromMap(map[string]interface{}{
	"Kind": "Service",
	"metadata": map[string]interface{}{
		"labels": map[string]interface{}{
			"app": "application-name",
		},
		"name": "service-name",
	},
	"spec": map[string]interface{}{
		"ports": map[string]interface{}{
			"port": int64(80),
		},
	},
	"this": map[string]interface{}{
		"is": map[string]interface{}{
			"aNumber":      int64(1000),
			"aFloat":       float64(1.001),
			"aNilValue":    nil,
			"aBool":        true,
			"anEmptyMap":   map[string]interface{}{},
			"anEmptySlice": []interface{}{},
			"unrecognizable": testing.InternalExample{
				Name: "fooBar",
			},
		},
	},
	"that": []interface{}{
		"idx0",
		"idx1",
		"idx2",
		"idx3",
	},
	"those": []interface{}{
		map[string]interface{}{
			"field1": "idx0foo",
			"field2": "idx0bar",
		},
		map[string]interface{}{
			"field1": "idx1foo",
			"field2": "idx1bar",
		},
		map[string]interface{}{
			"field1": "idx2foo",
			"field2": "idx2bar",
		},
	},
	"these": []interface{}{
		map[string]interface{}{
			"field1": []interface{}{"idx010", "idx011"},
			"field2": []interface{}{"idx020", "idx021"},
		},
		map[string]interface{}{
			"field1": []interface{}{"idx110", "idx111"},
			"field2": []interface{}{"idx120", "idx121"},
		},
		map[string]interface{}{
			"field1": []interface{}{"idx210", "idx211"},
			"field2": []interface{}{"idx220", "idx221"},
		},
	},
	"complextree": []interface{}{
		map[string]interface{}{
			"field1": []interface{}{
				map[string]interface{}{
					"stringsubfield": "idx1010",
					"intsubfield":    int64(1010),
					"floatsubfield":  float64(1.010),
					"boolfield":      true,
				},
				map[string]interface{}{
					"stringsubfield": "idx1011",
					"intsubfield":    int64(1011),
					"floatsubfield":  float64(1.011),
					"boolfield":      false,
				},
			},
			"field2": []interface{}{
				map[string]interface{}{
					"stringsubfield": "idx1020",
					"intsubfield":    int64(1020),
					"floatsubfield":  float64(1.020),
					"boolfield":      true,
				},
				map[string]interface{}{
					"stringsubfield": "idx1021",
					"intsubfield":    int64(1021),
					"floatsubfield":  float64(1.021),
					"boolfield":      false,
				},
			},
		},
		map[string]interface{}{
			"field1": []interface{}{
				map[string]interface{}{
					"stringsubfield": "idx1110",
					"intsubfield":    int64(1110),
					"floatsubfield":  float64(1.110),
					"boolfield":      true,
				},
				map[string]interface{}{
					"stringsubfield": "idx1111",
					"intsubfield":    int64(1111),
					"floatsubfield":  float64(1.111),
					"boolfield":      false,
				},
			},
			"field2": []interface{}{
				map[string]interface{}{
					"stringsubfield": "idx1120",
					"intsubfield":    int64(1120),
					"floatsubfield":  float64(1.1120),
					"boolfield":      true,
				},
				map[string]interface{}{
					"stringsubfield": "idx1121",
					"intsubfield":    int64(1121),
					"floatsubfield":  float64(1.1121),
					"boolfield":      false,
				},
			},
		},
	},
})

func TestGetFieldValue(t *testing.T) {
	tests := []struct {
		name          string
		pathToField   string
		expectedValue interface{}
		errorExpected bool
		errorMsg      string
	}{
		{
			name:          "oneField",
			pathToField:   "Kind",
			expectedValue: "Service",
			errorExpected: false,
		},
		{
			name:          "twoFields",
			pathToField:   "metadata.name",
			expectedValue: "service-name",
			errorExpected: false,
		},
		{
			name:          "threeFields",
			pathToField:   "spec.ports.port",
			expectedValue: int64(80),
			errorExpected: false,
		},
		{
			name:          "empty",
			pathToField:   "",
			errorExpected: true,
			errorMsg:      "no field named ''",
		},
		{
			name:          "emptyDotEmpty",
			pathToField:   ".",
			errorExpected: true,
			errorMsg:      "no field named '.'",
		},
		{
			name:          "twoFieldsOneMissing",
			pathToField:   "metadata.banana",
			errorExpected: true,
			errorMsg:      "no field named 'metadata.banana'",
		},
		{
			name:          "deeperMissingField",
			pathToField:   "this.is.aDeep.field.that.does.not.exist",
			errorExpected: true,
			errorMsg:      "no field named 'this.is.aDeep.field.that.does.not.exist'",
		},
		{
			name:          "emptyMap",
			pathToField:   "this.is.anEmptyMap",
			errorExpected: false,
			expectedValue: map[string]interface{}{},
		},
		{
			name:          "emptySlice",
			pathToField:   "this.is.anEmptySlice",
			errorExpected: false,
			expectedValue: []interface{}{},
		},
		{
			name:          "numberAsValue",
			pathToField:   "this.is.aNumber",
			errorExpected: false,
			expectedValue: int64(1000),
		},
		{
			name:          "floatAsValue",
			pathToField:   "this.is.aFloat",
			errorExpected: false,
			expectedValue: float64(1.001),
		},
		{
			name:          "boolAsValue",
			pathToField:   "this.is.aBool",
			errorExpected: false,
			expectedValue: true,
		},
		{
			name:          "nilAsValue",
			pathToField:   "this.is.aNilValue",
			errorExpected: false,
			expectedValue: nil,
		},
		{
			name:          "unrecognizable",
			pathToField:   "this.is.unrecognizable.Name",
			errorExpected: true,
			errorMsg:      ".this.is.unrecognizable.Name accessor error: {fooBar <nil>  false} is of the type testing.InternalExample, expected map[string]interface{}",
		},
		{
			name:          "validStringIndex",
			pathToField:   "that[2]",
			expectedValue: "idx2",
			errorExpected: false,
		},
		{
			name:          "outOfBoundIndex",
			pathToField:   "that[99]",
			errorMsg:      "no field named 'that[99]'",
			errorExpected: true,
		},
		{
			name:          "accessorError",
			pathToField:   "that[downwardapi]",
			errorMsg:      ".that.downwardapi accessor error: [idx0 idx1 idx2 idx3] is of the type []interface {}, expected map[string]interface{}",
			errorExpected: true,
		},
		{
			name:          "unknownSlice",
			pathToField:   "unknown[0]",
			errorMsg:      "no field named 'unknown[0]'",
			errorExpected: true,
		},
		{
			name:          "sliceInSlice",
			pathToField:   "that[2][0]",
			errorExpected: true,
			errorMsg:      "no field named 'that[2][0]'",
		},
		{
			name:          "validStructIndex",
			pathToField:   "those[1]",
			errorExpected: false,
			expectedValue: map[string]interface{}{"field1": "idx1foo", "field2": "idx1bar"},
		},
		{
			name:          "validStructSubField",
			pathToField:   "those[1].field2",
			errorExpected: false,
			expectedValue: "idx1bar",
		},
		{
			name:          "validStructSubFieldIndex",
			pathToField:   "these[1].field2[1]",
			errorExpected: false,
			expectedValue: "idx121",
		},
		{
			name:          "validStructSubFieldOutOfBoundIndex",
			pathToField:   "these[1].field2[99]",
			errorExpected: true,
			errorMsg:      "no field named 'these[1].field2[99]'",
		},
		{
			name:          "validStructSubFieldIndexSubfield",
			pathToField:   "complextree[1].field2[1].stringsubfield",
			errorExpected: false,
			expectedValue: "idx1121",
		},
		{
			name:          "validStructSubFieldIndexInvalidName",
			pathToField:   "complextree[1].field2[1].invalidsubfield",
			errorExpected: true,
			errorMsg:      "no field named 'complextree[1].field2[1].invalidsubfield'",
		},
		{
			name:          "validDownwardAPILabels",
			pathToField:   `metadata.labels["app"]`,
			errorExpected: false,
			expectedValue: "application-name",
		},
		{
			name:          "validDownwardAPISpecs",
			pathToField:   `spec.ports['port']`,
			errorExpected: false,
			expectedValue: int64(80),
		},
		{
			name:          "validDownwardAPIThis",
			pathToField:   `this.is[aFloat]`,
			errorExpected: false,
			expectedValue: float64(1.001),
		},
		{
			name:          "downwardAPIInvalidLabel",
			pathToField:   `metadata.labels["theisnotanint"]`,
			errorExpected: true,
			errorMsg:      `no field named 'metadata.labels["theisnotanint"]'`,
		},
		{
			name:          "downwardAPIInvalidLabel2",
			pathToField:   `invalidfield.labels["app"]`,
			errorExpected: true,
			errorMsg:      `no field named 'invalidfield.labels["app"]'`,
		},
		{
			name:          "invalidIndexInIndex",
			pathToField:   "complextree[1[0]]",
			errorExpected: true,
			errorMsg:      "no field named 'complextree[1[0]]'",
		},
		{
			name:          "invalidClosingBrackets",
			pathToField:   "complextree[1]]",
			errorExpected: true,
			errorMsg:      "no field named 'complextree[1]]'",
		},
		{
			name:          "validFieldsWithQuotes",
			pathToField:   "'complextree'[1].field2[1].'stringsubfield'",
			errorExpected: false,
			expectedValue: "idx1121",
		},
	}

	for _, test := range tests {
		s, err := kunstructured.GetFieldValue(test.pathToField)
		if test.errorExpected {
			compareExpectedError(t, test.name, test.pathToField, err, test.errorMsg)
			continue
		}
		if err != nil {
			unExpectedError(t, test.name, test.pathToField, err)
		}
		compareValues(t, test.name, test.pathToField, test.expectedValue, s)
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name          string
		pathToField   string
		expectedValue string
		errorExpected bool
		errorMsg      string
	}{
		{
			name:          "oneField",
			pathToField:   "Kind",
			expectedValue: "Service",
			errorExpected: false,
		},
		{
			name:          "twoFields",
			pathToField:   "metadata.name",
			expectedValue: "service-name",
			errorExpected: false,
		},
		{
			name:          "emptyMap",
			pathToField:   "this.is.anEmptyMap",
			errorExpected: true,
			errorMsg:      ".this.is.anEmptyMap accessor error: map[] is of the type map[string]interface {}, expected string",
		},
		{
			name:          "twoFieldsOneMissing",
			pathToField:   "metadata.banana",
			errorExpected: true,
			errorMsg:      "no field named 'metadata.banana'",
		},
		{
			name:          "emptySlice",
			pathToField:   "this.is.anEmptySlice",
			errorExpected: true,
			errorMsg:      ".this.is.anEmptySlice accessor error: [] is of the type []interface {}, expected string",
		},
		{
			name:          "numberAsValue",
			pathToField:   "this.is.aNumber",
			errorExpected: true,
			errorMsg:      ".this.is.aNumber accessor error: 1000 is of the type int64, expected string",
		},
		{
			name:          "nilAsValue",
			pathToField:   "this.is.aNilValue",
			errorExpected: true,
			errorMsg:      ".this.is.aNilValue accessor error: <nil> is of the type <nil>, expected string",
		},
		{
			name:          "validStringIndex",
			pathToField:   "that[2]",
			expectedValue: "idx2",
			errorExpected: false,
		},
		{
			name:          "validStructIndex",
			pathToField:   "those[1]",
			errorExpected: true,
			errorMsg:      ".[1] accessor error: map[field1:idx1foo field2:idx1bar] is of the type map[string]interface {}, expected string",
		},
		{
			name:          "validStructSubField",
			pathToField:   "those[1].field2",
			errorExpected: false,
			expectedValue: "idx1bar",
		},
		{
			name:          "validStructSubFieldIndex",
			pathToField:   "these[1].field2[1]",
			errorExpected: false,
			expectedValue: "idx121",
		},
		{
			name:          "validStructSubFieldIndexSubfield",
			pathToField:   "complextree[1].field2[1].stringsubfield",
			errorExpected: false,
			expectedValue: "idx1121",
		},
		{
			name:          "invalidIndexInMap",
			pathToField:   "this.is[1]",
			errorExpected: true,
			errorMsg:      "no field named 'this.is[1]'",
		},
		{
			name:          "anotherInvalidIndexInMap",
			pathToField:   "this.is[1].aString",
			errorExpected: true,
			errorMsg:      "no field named 'this.is[1].aString'",
		},
		{
			name:          "validDownwardAPIField",
			pathToField:   `metadata.labels["app"]`,
			errorExpected: false,
			expectedValue: "application-name",
		},
	}

	for _, test := range tests {
		s, err := kunstructured.GetString(test.pathToField)
		if test.errorExpected {
			compareExpectedError(t, test.name, test.pathToField, err, test.errorMsg)
			continue
		}
		if err != nil {
			unExpectedError(t, test.name, test.pathToField, err)
		}
		compareValues(t, test.name, test.pathToField, test.expectedValue, s)
	}
}

func TestGetSlice(t *testing.T) {
	tests := []struct {
		name          string
		pathToField   string
		errorExpected bool
		errorMsg      string
	}{
		{
			name:          "validSlice1",
			pathToField:   "that",
			errorExpected: false,
		},
		{
			name:          "validSlice2",
			pathToField:   "those",
			errorExpected: false,
		},
		{
			name:          "validSlice3",
			pathToField:   "these",
			errorExpected: false,
		},
		{
			name:          "validSlice4",
			pathToField:   "complextree",
			errorExpected: false,
		},
		{
			name:          "twoFieldsOneMissing",
			pathToField:   "metadata.banana",
			errorExpected: true,
			errorMsg:      "no field named 'metadata.banana'",
		},
		{
			name:          "validStructSubFieldOutOfBoundIndex",
			pathToField:   "these[1].field2[99]",
			errorExpected: true,
			errorMsg:      "no field named 'these[1].field2[99]'",
		},
	}

	for _, test := range tests {
		_, err := kunstructured.GetSlice(test.pathToField)
		if test.errorExpected {
			compareExpectedError(t, test.name, test.pathToField, err, test.errorMsg)
			continue
		}
		if err != nil {
			unExpectedError(t, test.name, test.pathToField, err)
		}
	}
}

// unExpectedError function handles unexpected error
func unExpectedError(t *testing.T, name string, pathToField string, err error) {
	t.Fatalf("%q; path %q - unexpected error %v", name, pathToField, err)
}

// compareExpectedError compares the expectedError and the actualError return by GetFieldValue
func compareExpectedError(t *testing.T, name string, pathToField string, err error, errorMsg string) {
	if err == nil {
		t.Fatalf("%q; path %q - should return error, but no error returned",
			name, pathToField)
	}

	if errorMsg != err.Error() {
		t.Fatalf("%q; path %q - expected error: \"%s\", got error: \"%v\"",
			name, pathToField, errorMsg, err.Error())
	}
}

// compareValues compares the expectedValue and actualValue returned by GetFieldValue
func compareValues(t *testing.T, name string, pathToField string, expectedValue interface{}, actualValue interface{}) {
	t.Helper()
	switch typedV := expectedValue.(type) {
	case nil, string, bool, float64, int, int64:
		if expectedValue != actualValue {
			t.Fatalf("%q; Got: %v Expected: %v", name, actualValue, expectedValue)
		}
	case map[string]interface{}:
		if !reflect.DeepEqual(expectedValue, actualValue) {
			t.Fatalf("%q; Got: %v Expected: %v", name, actualValue, expectedValue)
		}
	case []interface{}:
		if !reflect.DeepEqual(expectedValue, actualValue) {
			t.Fatalf("%q; Got: %v Expected: %v", name, actualValue, expectedValue)
		}
	default:
		t.Logf("%T value at `%s`", typedV, pathToField)
	}
}
