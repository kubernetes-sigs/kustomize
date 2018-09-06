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

package app

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/internal/loadertest"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	kustomizationContent1 = `
namePrefix: foo-
namespace: ns1
commonLabels:
  app: nginx
commonAnnotations:
  note: This is a test annotation
resources:
  - deployment.yaml
  - namespace.yaml
configMapGenerator:
- name: literalConfigMap
  literals:
  - DB_USERNAME=admin
  - DB_PASSWORD=somepw
secretGenerator:
- name: secret
  commands:
    DB_USERNAME: "printf admin"
    DB_PASSWORD: "printf somepw"
  type: Opaque
patchesJson6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: dply1
  path: jsonpatch.json
`
	kustomizationContent2 = `
secretGenerator:
- name: secret
  timeoutSeconds: 1
  commands:
    USER: "sleep 2"
  type: Opaque
`
	deploymentContent = `apiVersion: apps/v1
metadata:
  name: dply1
kind: Deployment
`
	namespaceContent = `apiVersion: v1
kind: Namespace
metadata:
  name: ns1
`
	jsonpatchContent = `[
    {"op": "add", "path": "/spec/replica", "value": "3"}
]`
)

func makeLoader1(t *testing.T) loader.Loader {
	ldr := loadertest.NewFakeLoader("/testpath")
	err := ldr.AddFile("/testpath/"+constants.KustomizationFileName, []byte(kustomizationContent1))
	if err != nil {
		t.Fatalf("Failed to setup fake ldr.")
	}
	err = ldr.AddFile("/testpath/deployment.yaml", []byte(deploymentContent))
	if err != nil {
		t.Fatalf("Failed to setup fake ldr.")
	}
	err = ldr.AddFile("/testpath/namespace.yaml", []byte(namespaceContent))
	if err != nil {
		t.Fatalf("Failed to setup fake ldr.")
	}
	err = ldr.AddFile("/testpath/jsonpatch.json", []byte(jsonpatchContent))
	if err != nil {
		t.Fatalf("Failed to setup fake ldr.")
	}
	return ldr
}

var deploy = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
var cmap = schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}
var secret = schema.GroupVersionKind{Version: "v1", Kind: "Secret"}
var ns = schema.GroupVersionKind{Version: "v1", Kind: "Namespace"}

func TestResources1(t *testing.T) {
	expected := resmap.ResMap{
		resource.NewResIdWithPrefixNamespace(deploy, "dply1", "foo-", "ns1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "foo-dply1",
					"namespace": "ns1",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
				},
				"spec": map[string]interface{}{
					"replica": "3",
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "nginx",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"note": "This is a test annotation",
							},
							"labels": map[string]interface{}{
								"app": "nginx",
							},
						},
					},
				},
			}),
		resource.NewResIdWithPrefixNamespace(cmap, "literalConfigMap", "foo-", "ns1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "foo-literalConfigMap-mc92bgcbh5",
					"namespace": "ns1",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
					"creationTimestamp": nil,
				},
				"data": map[string]interface{}{
					"DB_USERNAME": "admin",
					"DB_PASSWORD": "somepw",
				},
			}).SetBehavior(resource.BehaviorCreate),
		resource.NewResIdWithPrefixNamespace(secret, "secret", "foo-", "ns1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name":      "foo-secret-877fcfhgt5",
					"namespace": "ns1",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
					"creationTimestamp": nil,
				},
				"type": string(corev1.SecretTypeOpaque),
				"data": map[string]interface{}{
					"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
					"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
				},
			}).SetBehavior(resource.BehaviorCreate),
		resource.NewResIdWithPrefixNamespace(ns, "ns1", "foo-", ""): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "foo-ns1",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
				},
			}),
	}
	l := makeLoader1(t)
	fakeFs := fs.MakeFakeFS()
	fakeFs.Mkdir("/")
	app, err := NewApplication(l, fakeFs)
	if err != nil {
		t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := app.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Unexpected Resources error %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		err = expected.ErrorIfNotEqual(actual)
		t.Fatalf("unexpected inequality: %v", err)
	}
}

func TestResourceNotFound(t *testing.T) {
	l := loadertest.NewFakeLoader("/testpath")
	err := l.AddFile("/testpath/"+constants.KustomizationFileName, []byte(kustomizationContent1))
	if err != nil {
		t.Fatalf("Failed to setup fake ldr.")
	}
	fakeFs := fs.MakeFakeFS()
	fakeFs.Mkdir("/")
	app, err := NewApplication(l, fakeFs)
	if err != nil {
		t.Fatalf("Unexpected construction error %v", err)
	}
	_, err = app.MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("Didn't get the expected error for an unknown resource")
	}
	if !strings.Contains(err.Error(), `cannot read file "/testpath/deployment.yaml"`) {
		t.Fatalf("Unpexpected error message %q", err)
	}
}

func TestSecretTimeout(t *testing.T) {
	l := loadertest.NewFakeLoader("/testpath")
	err := l.AddFile("/testpath/"+constants.KustomizationFileName, []byte(kustomizationContent2))
	if err != nil {
		t.Fatalf("Failed to setup fake ldr.")
	}
	fakeFs := fs.MakeFakeFS()
	fakeFs.Mkdir("/")
	app, err := NewApplication(l, fakeFs)
	if err != nil {
		t.Fatalf("Unexpected construction error %v", err)
	}
	_, err = app.MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("Didn't get the expected error for an unknown resource")
	}
	if !strings.Contains(err.Error(), "killed") {
		t.Fatalf("Unpexpected error message %q", err)
	}
}
