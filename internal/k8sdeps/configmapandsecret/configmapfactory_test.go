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

package configmapandsecret

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/types"
)

func makeEnvConfigMap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"DB_USERNAME": "admin",
			"DB_PASSWORD": "somepw",
		},
	}
}

func makeFileConfigMap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"app-init.ini": `FOO=bar
BAR=baz
`,
		},
	}
}

func makeLiteralConfigMap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"a": "x",
			"b": "y",
			"c": "Hello World",
			"d": "true",
		},
	}
}

func TestConstructConfigMap(t *testing.T) {
	type testCase struct {
		description string
		input       types.ConfigMapArgs
		expected    *corev1.ConfigMap
	}

	testCases := []testCase{
		{
			description: "construct config map from env",
			input: types.ConfigMapArgs{
				Name: "envConfigMap",
				DataSources: types.DataSources{
					EnvSource: "configmap/app.env",
				},
			},
			expected: makeEnvConfigMap("envConfigMap"),
		},
		{
			description: "construct config map from file",
			input: types.ConfigMapArgs{
				Name: "fileConfigMap",
				DataSources: types.DataSources{
					FileSources: []string{"configmap/app-init.ini"},
				},
			},
			expected: makeFileConfigMap("fileConfigMap"),
		},
		{
			description: "construct config map from literal",
			input: types.ConfigMapArgs{
				Name: "literalConfigMap",
				DataSources: types.DataSources{
					LiteralSources: []string{"a=x", "b=y", "c=\"Hello World\"", "d='true'"},
				},
			},
			expected: makeLiteralConfigMap("literalConfigMap"),
		},
	}

	fSys := fs.MakeFakeFS()
	fSys.WriteFile("configmap/app.env", []byte("DB_USERNAME=admin\nDB_PASSWORD=somepw\n"))
	fSys.WriteFile("configmap/app-init.ini", []byte("FOO=bar\nBAR=baz\n"))
	f := NewConfigMapFactory(fSys, loader.NewFileLoader(fSys))
	for _, tc := range testCases {
		cm, err := f.MakeConfigMap(&tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(*cm, *tc.expected) {
			t.Fatalf("in testcase: %q updated:\n%#v\ndoesn't match expected:\n%#v\n", tc.description, *cm, tc.expected)
		}
	}
}
