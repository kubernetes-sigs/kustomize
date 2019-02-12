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

package resource_test

import (
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	. "sigs.k8s.io/kustomize/pkg/resource"
)

var factory = NewFactory(
	kunstruct.NewKunstructuredFactoryImpl())

var testConfigMap = factory.FromMap(
	map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "winnie",
			"namespace": "hundred-acre-wood",
		},
	})

const genArgOptions = "{nsfx:false,beh:unspecified}"

const configMapAsString = `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie","namespace":"hundred-acre-wood"}}`

var testDeployment = factory.FromMap(
	map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "pooh",
		},
	})

const deploymentAsString = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"pooh"}}`

func TestResourceString(t *testing.T) {
	tests := []struct {
		in *Resource
		s  string
	}{
		{
			in: testConfigMap,
			s:  configMapAsString + genArgOptions,
		},
		{
			in: testDeployment,
			s:  deploymentAsString + genArgOptions,
		},
	}
	for _, test := range tests {
		if test.in.String() != test.s {
			t.Fatalf("Expected %s == %s", test.in.String(), test.s)
		}
	}
}

func TestResourceId(t *testing.T) {
	tests := []struct {
		in *Resource
		id resid.ResId
	}{
		{
			in: testConfigMap,
			id: resid.NewResIdWithPrefixNamespace(gvk.Gvk{Version: "v1", Kind: "ConfigMap"}, "winnie", "", "hundred-acre-wood"),
		},
		{
			in: testDeployment,
			id: resid.NewResId(gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}, "pooh"),
		},
	}
	for _, test := range tests {
		if test.in.Id() != test.id {
			t.Fatalf("Expected %v, but got %v\n", test.id, test.in.Id())
		}
	}
}
