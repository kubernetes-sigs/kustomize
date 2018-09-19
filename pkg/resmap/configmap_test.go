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

package resmap

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/pkg/configmapandsecret"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/internal/loadertest"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
)

var cmap = schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}

func TestNewFromConfigMaps(t *testing.T) {
	type testCase struct {
		description string
		input       []types.ConfigMapArgs
		filepath    string
		content     string
		expected    ResMap
	}

	l := loadertest.NewFakeLoader("/home/seans/project/")
	f := configmapandsecret.NewConfigMapFactory(fs.MakeFakeFS(), l)
	testCases := []testCase{
		{
			description: "construct config map from env",
			input: []types.ConfigMapArgs{
				{
					Name: "envConfigMap",
					DataSources: types.DataSources{
						EnvSource: "app.env",
					},
				},
			},
			filepath: "/home/seans/project/app.env",
			content:  "DB_USERNAME=admin\nDB_PASSWORD=somepw",
			expected: ResMap{
				resource.NewResId(cmap, "envConfigMap"): resource.NewResourceFromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "envConfigMap",
							"annotations": map[string]interface{}{
								"kustomize.sigs.k8s.io/generated": "true",
							},
							"creationTimestamp": nil,
						},
						"data": map[string]interface{}{
							"DB_USERNAME": "admin",
							"DB_PASSWORD": "somepw",
						},
					}).SetBehavior(resource.BehaviorCreate),
			},
		},
		{
			description: "construct config map from file",
			input: []types.ConfigMapArgs{{
				Name: "fileConfigMap",
				DataSources: types.DataSources{
					FileSources: []string{"app-init.ini"},
				},
			},
			},
			filepath: "/home/seans/project/app-init.ini",
			content:  "FOO=bar\nBAR=baz\n",
			expected: ResMap{
				resource.NewResId(cmap, "fileConfigMap"): resource.NewResourceFromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "fileConfigMap",
							"annotations": map[string]interface{}{
								"kustomize.sigs.k8s.io/generated": "true",
							},
							"creationTimestamp": nil,
						},
						"data": map[string]interface{}{
							"app-init.ini": `FOO=bar
BAR=baz
`,
						},
					}).SetBehavior(resource.BehaviorCreate),
			},
		},
		{
			description: "construct config map from literal",
			input: []types.ConfigMapArgs{
				{
					Name: "literalConfigMap",
					DataSources: types.DataSources{
						LiteralSources: []string{"a=x", "b=y", "c=\"Good Morning\"", "d=\"false\""},
					},
				},
			},
			expected: ResMap{
				resource.NewResId(cmap, "literalConfigMap"): resource.NewResourceFromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "literalConfigMap",
							"annotations": map[string]interface{}{
								"kustomize.sigs.k8s.io/generated": "true",
							},
							"creationTimestamp": nil,
						},
						"data": map[string]interface{}{
							"a": "x",
							"b": "y",
							"c": "Good Morning",
							"d": "false",
						},
					}).SetBehavior(resource.BehaviorCreate),
			},
		},
		// TODO: add testcase for data coming from multiple sources like
		// files/literal/env etc.
	}

	for _, tc := range testCases {
		if ferr := l.AddFile(tc.filepath, []byte(tc.content)); ferr != nil {
			t.Fatalf("Error adding fake file: %v\n", ferr)
		}
		r, err := NewResMapFromConfigMapArgs(f, tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(r, tc.expected) {
			t.Fatalf("in testcase: %q got:\n%+v\n expected:\n%+v\n", tc.description, r, tc.expected)
		}
	}
}
