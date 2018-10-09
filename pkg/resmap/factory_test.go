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
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/internal/loadertest"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/types"
)

func TestFromFiles(t *testing.T) {

	resourceStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: dply1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dply2
---
# some comment
---
---
`

	l := loadertest.NewFakeLoader("/home/seans/project")
	if ferr := l.AddFile("/home/seans/project/deployment.yaml", []byte(resourceStr)); ferr != nil {
		t.Fatalf("Error adding fake file: %v\n", ferr)
	}
	expected := ResMap{resid.NewResId(deploy, "dply1"): rf.FromMap(
		map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "dply1",
			},
		}),
		resid.NewResId(deploy, "dply2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "dply2",
				},
			}),
	}

	m, _ := rmF.FromFiles(
		l, []string{"/home/seans/project/deployment.yaml"})
	if len(m) != 2 {
		t.Fatalf("%#v should contain 2 appResource, but got %d", m, len(m))
	}

	if err := expected.ErrorIfNotEqual(m); err != nil {
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestFromBytes(t *testing.T) {
	encoded := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
`)
	expected := ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
		resid.NewResId(cmap, "cm2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm2",
				},
			}),
	}
	m, err := rmF.newResMapFromBytes(encoded)
	fmt.Printf("%v\n", m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		t.Fatalf("%#v doesn't match expected %#v", m, expected)
	}
}

var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}

func TestNewFromConfigMaps(t *testing.T) {
	type testCase struct {
		description string
		input       []types.ConfigMapArgs
		filepath    string
		content     string
		expected    ResMap
	}

	l := loadertest.NewFakeLoader("/home/seans/project/")
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
				resid.NewResId(cmap, "envConfigMap"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "envConfigMap",
						},
						"data": map[string]interface{}{
							"DB_USERNAME": "admin",
							"DB_PASSWORD": "somepw",
						},
					}).SetBehavior(ifc.BehaviorCreate),
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
				resid.NewResId(cmap, "fileConfigMap"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "fileConfigMap",
						},
						"data": map[string]interface{}{
							"app-init.ini": `FOO=bar
BAR=baz
`,
						},
					}).SetBehavior(ifc.BehaviorCreate),
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
				resid.NewResId(cmap, "literalConfigMap"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "literalConfigMap",
						},
						"data": map[string]interface{}{
							"a": "x",
							"b": "y",
							"c": "Good Morning",
							"d": "false",
						},
					}).SetBehavior(ifc.BehaviorCreate),
			},
		},
		// TODO: add testcase for data coming from multiple sources like
		// files/literal/env etc.
	}
	rmF.Set(fs.MakeFakeFS(), l)
	for _, tc := range testCases {
		if ferr := l.AddFile(tc.filepath, []byte(tc.content)); ferr != nil {
			t.Fatalf("Error adding fake file: %v\n", ferr)
		}
		r, err := rmF.NewResMapFromConfigMapArgs(tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(r, tc.expected) {
			t.Fatalf("in testcase: %q got:\n%+v\n expected:\n%+v\n", tc.description, r, tc.expected)
		}
	}
}

var secret = gvk.Gvk{Version: "v1", Kind: "Secret"}

func TestNewResMapFromSecretArgs(t *testing.T) {
	secrets := []types.SecretArgs{
		{
			Name: "apple",
			CommandSources: types.CommandSources{
				Commands: map[string]string{
					"DB_USERNAME": "printf admin",
					"DB_PASSWORD": "printf somepw",
				},
			},
			Type: ifc.SecretTypeOpaque,
		},
		{
			Name: "peanuts",
			CommandSources: types.CommandSources{
				EnvCommand: "printf \"DB_USERNAME=admin\nDB_PASSWORD=somepw\"",
			},
			Type: ifc.SecretTypeOpaque,
		},
	}
	fakeFs := fs.MakeFakeFS()
	fakeFs.Mkdir(".")
	rmF.Set(fakeFs, loader.NewFileLoader(fakeFs))
	actual, err := rmF.NewResMapFromSecretArgs(secrets)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := ResMap{
		resid.NewResId(secret, "apple"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name": "apple",
				},
				"type": ifc.SecretTypeOpaque,
				"data": map[string]interface{}{
					"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
					"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
				},
			}).SetBehavior(ifc.BehaviorCreate),
		resid.NewResId(secret, "peanuts"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name": "peanuts",
				},
				"type": ifc.SecretTypeOpaque,
				"data": map[string]interface{}{
					"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
					"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
				},
			}).SetBehavior(ifc.BehaviorCreate),
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%#v\ndoesn't match expected:\n%#v", actual, expected)
	}
}

func TestSecretTimeout(t *testing.T) {
	timeout := int64(1)
	secrets := []types.SecretArgs{
		{
			Name:           "slow",
			TimeoutSeconds: &timeout,
			CommandSources: types.CommandSources{
				Commands: map[string]string{
					"USER": "sleep 2",
				},
			},
			Type: ifc.SecretTypeOpaque,
		},
	}
	fakeFs := fs.MakeFakeFS()
	fakeFs.Mkdir(".")
	rmF.Set(fakeFs, loader.NewFileLoader(fakeFs))
	_, err := rmF.NewResMapFromSecretArgs(secrets)

	if err == nil {
		t.Fatal("didn't get the expected timeout error", err)
	}
}
