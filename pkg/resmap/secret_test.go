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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/pkg/configmapandsecret"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
)

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
			Type: "Opaque",
		},
		{
			Name: "peanuts",
			CommandSources: types.CommandSources{
				EnvCommand: "printf \"DB_USERNAME=admin\nDB_PASSWORD=somepw\"",
			},
			Type: "Opaque",
		},
	}
	fakeFs := fs.MakeFakeFS()
	fakeFs.Mkdir(".")
	actual, err := NewResMapFromSecretArgs(
		configmapandsecret.NewSecretFactory(fakeFs, "."), secrets)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := ResMap{
		resource.NewResId(secret, "apple"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name": "apple",
				},
				"type": string(corev1.SecretTypeOpaque),
				"data": map[string]interface{}{
					"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
					"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
				},
			}).SetBehavior(resource.BehaviorCreate),
		resource.NewResId(secret, "peanuts"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name": "peanuts",
				},
				"type": string(corev1.SecretTypeOpaque),
				"data": map[string]interface{}{
					"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
					"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
				},
			}).SetBehavior(resource.BehaviorCreate),
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
			Type: "Opaque",
		},
	}
	fakeFs := fs.MakeFakeFS()
	fakeFs.Mkdir(".")
	_, err := NewResMapFromSecretArgs(
		configmapandsecret.NewSecretFactory(fakeFs, "."), secrets)

	if err == nil {
		t.Fatal("didn't get the expected timeout error", err)
	}
}
