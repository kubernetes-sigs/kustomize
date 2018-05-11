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
	"encoding/base64"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kubectl/pkg/kustomize/types"
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

func makeUnstructuredEnvConfigMap(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":              name,
				"creationTimestamp": nil,
			},
			"data": map[string]interface{}{
				"DB_USERNAME": "admin",
				"DB_PASSWORD": "somepw",
			},
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
		},
	}
}

func makeTestSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			"DB_USERNAME": []byte("admin"),
			"DB_PASSWORD": []byte("somepw"),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

func makeUnstructuredSecret(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":              name,
				"creationTimestamp": nil,
			},
			"type": string(corev1.SecretTypeOpaque),
			"data": map[string]interface{}{
				"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
				"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
			},
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
					EnvSource: "../examples/simple/instances/exampleinstance/configmap/app.env",
				},
			},
			expected: makeEnvConfigMap("envConfigMap"),
		},
		{
			description: "construct config map from file",
			input: types.ConfigMapArgs{
				Name: "fileConfigMap",
				DataSources: types.DataSources{
					FileSources: []string{"../examples/simple/instances/exampleinstance/configmap/app-init.ini"},
				},
			},
			expected: makeFileConfigMap("fileConfigMap"),
		},
		{
			description: "construct config map from literal",
			input: types.ConfigMapArgs{
				Name: "literalConfigMap",
				DataSources: types.DataSources{
					LiteralSources: []string{"a=x", "b=y"},
				},
			},
			expected: makeLiteralConfigMap("literalConfigMap"),
		},
	}

	for _, tc := range testCases {
		cm, err := makeConfigMap(tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(*cm, *tc.expected) {
			t.Fatalf("in testcase: %q updated:\n%#v\ndoesn't match expected:\n%#v\n", tc.description, *cm, tc.expected)
		}
	}
}

func TestConstructSecret(t *testing.T) {
	secret := types.SecretArgs{
		Name: "secret",
		Commands: map[string]string{
			"DB_USERNAME": "printf admin",
			"DB_PASSWORD": "printf somepw",
		},
		Type: "Opaque",
	}
	cm, err := makeSecret(secret, ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := makeTestSecret("secret")
	if !reflect.DeepEqual(*cm, *expected) {
		t.Fatalf("%#v\ndoesn't match expected:\n%#v", *cm, *expected)
	}
}

func TestFailConstructSecret(t *testing.T) {
	secret := types.SecretArgs{
		Name: "secret",
		Commands: map[string]string{
			"FAILURE": "false", // This will fail.
		},
		Type: "Opaque",
	}
	_, err := makeSecret(secret, ".")
	if err == nil {
		t.Fatalf("Expected failure.")
	}
}

func TestObjectConvertToUnstructured(t *testing.T) {
	type testCase struct {
		description string
		input       *corev1.ConfigMap
		expected    *unstructured.Unstructured
	}

	testCases := []testCase{
		{
			description: "convert config map",
			input:       makeEnvConfigMap("envConfigMap"),
			expected:    makeUnstructuredEnvConfigMap("envConfigMap"),
		},
		{
			description: "convert secret",
			input:       makeEnvConfigMap("envSecret"),
			expected:    makeUnstructuredEnvConfigMap("envSecret"),
		},
	}
	for _, tc := range testCases {
		actual, err := objectToUnstructured(tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Fatalf("%#v\ndoesn't match expected\n%#v\n", actual, tc.expected)
		}
	}
}
