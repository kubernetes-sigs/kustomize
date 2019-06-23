// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configmapandsecret

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
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
		BinaryData: map[string][]byte{
			"app.bin": {0xff, 0xfd},
		},
	}
}

func makeLiteralConfigMap(name string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
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
	cm.SetLabels(map[string]string{"foo": "bar"})
	return cm
}

func TestConstructConfigMap(t *testing.T) {
	type testCase struct {
		description string
		input       types.ConfigMapArgs
		options     *types.GeneratorOptions
		expected    *corev1.ConfigMap
	}

	testCases := []testCase{
		{
			description: "construct config map from env",
			input: types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "envConfigMap",
					DataSources: types.DataSources{
						EnvSources: []string{"configmap/app.env"},
					},
				},
			},
			options:  nil,
			expected: makeEnvConfigMap("envConfigMap"),
		},
		{
			description: "construct config map from file",
			input: types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "fileConfigMap",
					DataSources: types.DataSources{
						FileSources: []string{"configmap/app-init.ini", "configmap/app.bin"},
					},
				},
			},
			options:  nil,
			expected: makeFileConfigMap("fileConfigMap"),
		},
		{
			description: "construct config map from literal",
			input: types.ConfigMapArgs{
				GeneratorArgs: types.GeneratorArgs{
					Name: "literalConfigMap",
					DataSources: types.DataSources{
						LiteralSources: []string{"a=x", "b=y", "c=\"Hello World\"", "d='true'"},
					},
				},
			},
			options: &types.GeneratorOptions{
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			expected: makeLiteralConfigMap("literalConfigMap"),
		},
	}

	fSys := fs.MakeFakeFS()
	fSys.WriteFile("/configmap/app.env", []byte("DB_USERNAME=admin\nDB_PASSWORD=somepw\n"))
	fSys.WriteFile("/configmap/app-init.ini", []byte("FOO=bar\nBAR=baz\n"))
	fSys.WriteFile("/configmap/app.bin", []byte{0xff, 0xfd})
	ldr := loader.NewFileLoaderAtRoot(validators.MakeFakeValidator(), fSys)
	for _, tc := range testCases {
		f := NewFactory(ldr, tc.options)
		cm, err := f.MakeConfigMap(&tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(*cm, *tc.expected) {
			t.Fatalf("in testcase: %q updated:\n%#v\ndoesn't match expected:\n%#v\n", tc.description, *cm, tc.expected)
		}
	}
}
