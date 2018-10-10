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

	"sigs.k8s.io/kustomize/internal/k8sdeps"
)

var factory = NewFactory(
	k8sdeps.NewKunstructuredFactoryImpl())

var testConfigMap = factory.FromMap(
	map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	})

const testConfigMapString = `unspecified:{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie"}}`

var testDeployment = factory.FromMap(
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
