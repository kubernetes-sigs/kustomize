package k8sdeps

import (
	"reflect"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"testing"
)

var testConfigMap = NewKunstructuredFromMap(
	map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	})

func TestNewKunstructuredSliceFromBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectedOut []ifc.Kunstructured
		expectedErr bool
	}{
		{
			name:        "garbage",
			input:       []byte("garbageIn: garbageOut"),
			expectedOut: []ifc.Kunstructured{},
			expectedErr: true,
		},
		{
			name:        "noBytes",
			input:       []byte{},
			expectedOut: []ifc.Kunstructured{},
			expectedErr: false,
		},
		{
			name: "goodJson",
			input: []byte(`
{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie"}}
`),
			expectedOut: []ifc.Kunstructured{testConfigMap},
			expectedErr: false,
		},
		{
			name: "goodYaml1",
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			expectedOut: []ifc.Kunstructured{testConfigMap},
			expectedErr: false,
		},
		{
			name: "goodYaml2",
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`),
			expectedOut: []ifc.Kunstructured{testConfigMap, testConfigMap},
			expectedErr: false,
		},
		{
			name: "garbageInOneOfTwoObjects",
			input: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
---
WOOOOOOOOOOOOOOOOOOOOOOOOT:  woot
`),
			expectedOut: []ifc.Kunstructured{},
			expectedErr: true,
		},
	}

	for _, test := range tests {
		rs, err := NewKunstructuredSliceFromBytes(
			test.input, NewKustDecoder())
		if test.expectedErr && err == nil {
			t.Fatalf("%v: should return error", test.name)
		}
		if !test.expectedErr && err != nil {
			t.Fatalf("%v: unexpected error: %s", test.name, err)
		}
		if len(rs) != len(test.expectedOut) {
			t.Fatalf("%s: length mismatch %d != %d",
				test.name, len(rs), len(test.expectedOut))
		}
		for i := range rs {
			if !reflect.DeepEqual(test.expectedOut[i], rs[i]) {
				t.Fatalf("%s: Got: %v\nexpected:%v",
					test.name, test.expectedOut[i], rs[i])
			}
		}
	}
}

func TestGetFieldValue(t *testing.T) {
	funStruct := NewKunstructuredFromMap(map[string]interface{}{
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
		s, err := funStruct.GetFieldValue(test.pathToField)
		if test.errorExpected && err == nil {
			t.Fatalf("should return error, but no error returned")
		} else {
			if test.expectedValue != s {
				t.Fatalf("Got:%s expected:%s", s, test.expectedValue)
			}
		}
	}
}
