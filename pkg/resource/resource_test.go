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
	"reflect"
	"sigs.k8s.io/kustomize/pkg/internal/k8sdeps"
	"testing"

	"sigs.k8s.io/kustomize/pkg/internal/loadertest"
	"sigs.k8s.io/kustomize/pkg/patch"
)

var testConfigMap = NewResourceFromMap(
	map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	})

const testConfigMapString = `unspecified:{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie"}}`

var testDeployment = NewResourceFromMap(
	map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "pooh",
		},
	})

const testDeploymentString = `unspecified:{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"pooh"}}`

func TestResourceString(t *testing.T) {
	tests := []struct {
		in *Resource
		s  string
	}{
		{
			in: testConfigMap,
			s:  testConfigMapString,
		},
		{
			in: testDeployment,
			s:  testDeploymentString,
		},
	}
	for _, test := range tests {
		if test.in.String() != test.s {
			t.Fatalf("Expected %s == %s", test.in.String(), test.s)
		}
	}
}

func TestNewResourceSliceFromPatches(t *testing.T) {
	patchGood1 := patch.StrategicMerge("/foo/patch1.yaml")
	patch1 := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pooh
`
	patchGood2 := patch.StrategicMerge("/foo/patch2.yaml")
	patch2 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
---
# some comment
---
---
`
	patchBad := patch.StrategicMerge("/foo/patch3.yaml")
	patch3 := `
WOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOT: woot
`
	l := loadertest.NewFakeLoader("/foo")
	l.AddFile(string(patchGood1), []byte(patch1))
	l.AddFile(string(patchGood2), []byte(patch2))
	l.AddFile(string(patchBad), []byte(patch3))

	tests := []struct {
		name        string
		input       []patch.StrategicMerge
		expectedOut []*Resource
		expectedErr bool
	}{
		{
			name:        "happy",
			input:       []patch.StrategicMerge{patchGood1, patchGood2},
			expectedOut: []*Resource{testDeployment, testConfigMap},
			expectedErr: false,
		},
		{
			name:        "badFileName",
			input:       []patch.StrategicMerge{patchGood1, "doesNotExist"},
			expectedOut: []*Resource{},
			expectedErr: true,
		},
		{
			name:        "badData",
			input:       []patch.StrategicMerge{patchGood1, patchBad},
			expectedOut: []*Resource{},
			expectedErr: true,
		},
	}
	for _, test := range tests {
		rs, err := NewResourceSliceFromPatches(
			l, test.input, k8sdeps.NewKustDecoder())
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

func TestNewResourceSliceFromBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectedOut []*Resource
		expectedErr bool
	}{
		{
			name:        "garbage",
			input:       []byte("garbageIn: garbageOut"),
			expectedOut: []*Resource{},
			expectedErr: true,
		},
		{
			name:        "noBytes",
			input:       []byte{},
			expectedOut: []*Resource{},
			expectedErr: false,
		},
		{
			name: "goodJson",
			input: []byte(`
{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie"}}
`),
			expectedOut: []*Resource{testConfigMap},
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
			expectedOut: []*Resource{testConfigMap},
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
			expectedOut: []*Resource{testConfigMap, testConfigMap},
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
			expectedOut: []*Resource{},
			expectedErr: true,
		},
	}

	for _, test := range tests {
		rs, err := NewResourceSliceFromBytes(
			test.input, k8sdeps.NewKustDecoder())
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
