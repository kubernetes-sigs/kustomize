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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	manifest "k8s.io/kubectl/pkg/apis/manifest/v1alpha1"
	"k8s.io/kubectl/pkg/kustomize/resource"
	"k8s.io/kubectl/pkg/loader/loadertest"
)

func TestNewFromConfigMaps(t *testing.T) {
	type testCase struct {
		description string
		input       []manifest.ConfigMapArgs
		filepath    string
		content     string
		expected    resource.ResourceCollection
	}

	l := loadertest.NewFakeLoader("/home/seans/project/")
	testCases := []testCase{
		{
			description: "construct config map from env",
			input: []manifest.ConfigMapArgs{
				{
					Name: "envConfigMap",
					DataSources: manifest.DataSources{
						EnvSource: "app.env",
					},
				},
			},
			filepath: "/home/seans/project/app.env",
			content:  "DB_USERNAME=admin\nDB_PASSWORD=somepw",
			expected: resource.ResourceCollection{
				{
					GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
					Name: "envConfigMap",
				}: &resource.Resource{
					Data: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":              "envConfigMap",
								"creationTimestamp": nil,
							},
							"data": map[string]interface{}{
								"DB_USERNAME": "admin",
								"DB_PASSWORD": "somepw",
							},
						},
					},
				},
			},
		},
		{
			description: "construct config map from file",
			input: []manifest.ConfigMapArgs{{
				Name: "fileConfigMap",
				DataSources: manifest.DataSources{
					FileSources: []string{"app-init.ini"},
				},
			},
			},
			filepath: "/home/seans/project/app-init.ini",
			content:  "FOO=bar\nBAR=baz\n",
			expected: resource.ResourceCollection{
				{
					GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
					Name: "fileConfigMap",
				}: &resource.Resource{
					Data: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":              "fileConfigMap",
								"creationTimestamp": nil,
							},
							"data": map[string]interface{}{
								"app-init.ini": `FOO=bar
BAR=baz
`,
							},
						},
					},
				},
			},
		},
		{
			description: "construct config map from literal",
			input: []manifest.ConfigMapArgs{
				{
					Name: "literalConfigMap",
					DataSources: manifest.DataSources{
						LiteralSources: []string{"a=x", "b=y"},
					},
				},
			},
			expected: resource.ResourceCollection{
				{
					GVK:  schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
					Name: "literalConfigMap",
				}: &resource.Resource{
					Data: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":              "literalConfigMap",
								"creationTimestamp": nil,
							},
							"data": map[string]interface{}{
								"a": "x",
								"b": "y",
							},
						},
					},
				},
			},
		},
		// TODO: add testcase for data coming from multiple sources like
		// files/literal/env etc.
	}

	for _, tc := range testCases {

		if ferr := l.AddFile(tc.filepath, []byte(tc.content)); ferr != nil {
			t.Fatalf("Error adding fake file: %v\n", ferr)
		}
		r, err := resource.NewFromConfigMaps(l, tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(r, tc.expected) {
			t.Fatalf("in testcase: %q got:\n%+v\n expected:\n%+v\n", tc.description, r, tc.expected)
		}
	}
}
