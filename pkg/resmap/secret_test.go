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

	"github.com/kubernetes-sigs/kustomize/pkg/configmapandsecret"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var secret = schema.GroupVersionKind{Version: "v1", Kind: "Secret"}

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
					"name":              "apple",
					"creationTimestamp": nil,
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
					"name":              "peanuts",
					"creationTimestamp": nil,
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
