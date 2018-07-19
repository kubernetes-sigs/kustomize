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

	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kubernetes-sigs/kustomize/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
					EnvSource: "../examplelayout/simple/instances/exampleinstance/configmap/app.env",
				},
			},
			expected: makeEnvConfigMap("envConfigMap"),
		},
		{
			description: "construct config map from file",
			input: types.ConfigMapArgs{
				Name: "fileConfigMap",
				DataSources: types.DataSources{
					FileSources: []string{"../examplelayout/simple/instances/exampleinstance/configmap/app-init.ini"},
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

func makeSecret(secret types.SecretArgs, path string) (*corev1.Secret, error) {
	corev1secret := &corev1.Secret{}
	corev1secret.APIVersion = "v1"
	corev1secret.Kind = "Secret"
	corev1secret.Name = secret.Name
	corev1secret.Type = corev1.SecretType(secret.Type)
	if corev1secret.Type == "" {
		corev1secret.Type = corev1.SecretTypeOpaque
	}
	corev1secret.Data = map[string][]byte{}

	for k, v := range secret.Commands {
		out, err := createSecretKey(path, v)
		if err != nil {
			return nil, err
		}
		corev1secret.Data[k] = out
	}

	return corev1secret, nil
}

func createSecretKey(wd string, command string) ([]byte, error) {
	fi, err := os.Stat(wd)
	if err != nil || !fi.IsDir() {
		wd = filepath.Dir(wd)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = wd

	return cmd.Output()
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
			t.Fatalf("%s: unexpected error: %v", tc.description, err)
		}
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Fatalf("%s: %#v\ndoesn't match expected\n%#v\n", tc.description, actual, tc.expected)
		}
	}
}
