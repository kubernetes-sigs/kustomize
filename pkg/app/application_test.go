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
	"os"
	"reflect"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
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
)

func makeLoader1(t *testing.T) loader.Loader {
	loader := loadertest.NewFakeLoader("/testpath")
	err := loader.AddFile("/testpath/"+constants.KustomizationFileName, []byte(kustomizationContent1))
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	err = loader.AddFile("/testpath/deployment.yaml", []byte(deploymentContent))
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	err = loader.AddFile("/testpath/namespace.yaml", []byte(namespaceContent))
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	return loader
}

var deploy = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
var cmap = schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}
var secret = schema.GroupVersionKind{Version: "v1", Kind: "Secret"}
var ns = schema.GroupVersionKind{Version: "v1", Kind: "Namespace"}
var svc = schema.GroupVersionKind{Version: "v1", Kind: "Service"}

func TestResources1(t *testing.T) {
	expected := resmap.ResMap{
		resource.NewResId(deploy, "dply1"): resource.NewResourceFromMap(
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
		resource.NewResId(cmap, "literalConfigMap"): resource.NewResourceFromMap(
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
			}),
		resource.NewResId(secret, "secret"): resource.NewResourceFromMap(
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
			}),
		resource.NewResId(ns, "ns1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name":      "foo-ns1",
					"namespace": "ns1",
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
	app, err := New(l)
	if err != nil {
		t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := app.Resources()
	if err != nil {
		t.Fatalf("Unexpected Resources error %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		err = expected.ErrorIfNotEqual(actual)
		t.Fatalf("unexpected inequality: %v", err)
	}
}

func TestRawResources1(t *testing.T) {
	expected := resmap.ResMap{
		resource.NewResId(deploy, "dply1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "dply1",
				},
			}),
		resource.NewResId(ns, "ns1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "ns1",
				},
			}),
	}
	l := makeLoader1(t)
	app, err := New(l)
	if err != nil {
		t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := app.RawResources()
	if err != nil {
		t.Fatalf("Unexpected RawResources error %v", err)
	}

	if err := expected.ErrorIfNotEqual(actual); err != nil {
		t.Fatalf("unexpected inequality: %v", err)
	}
}

const (
	kustomizationContentBase = `
namePrefix: foo-
commonLabels:
  app: banana
resources:
  - deployment.yaml
`
	kustomizationContentOverlay = `
commonLabels:
  env: staging
resources:
  - service.yaml
bases:
  - base
`
	serviceContent = `apiVersion: v1
kind: Service
metadata:
  name: svc
spec:
  type: LoadBalancer
`
)

func makeLoader2(t *testing.T) loader.Loader {
	loader := loadertest.NewFakeLoader("/testpath")
	err := loader.AddFile("/testpath/"+constants.KustomizationFileName, []byte(kustomizationContentOverlay))
	if err != nil {
		t.Fatal(err)
	}
	err = loader.AddFile("/testpath/service.yaml", []byte(serviceContent))
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	err = loader.AddDirectory("/testpath/base", os.ModeDir)
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	err = loader.AddFile("/testpath/base/"+constants.KustomizationFileName, []byte(kustomizationContentBase))
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	err = loader.AddFile("/testpath/base/deployment.yaml", []byte(deploymentContent))
	if err != nil {
		t.Fatalf("Failed to setup fake loader.")
	}
	return loader
}

// TODO: This test covers incorrect behavior; it should not pass.
// It asks for raw resources.  The Service resource is returned in raw form,
// but the resources in the base are modified to have the banana label,
// the 'foo' name prefix, etc.  This method exists only to support the
// diff command comparing customized to non-customized resources;
// perhaps it's not worth supporting the command.
func TestRawResources2(t *testing.T) {
	expected := resmap.ResMap{
		resource.NewResId(deploy, "dply1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "foo-dply1",
					"labels": map[string]interface{}{
						"app": "banana",
					},
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "banana",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "banana",
							},
						},
					},
				},
			}),
		resource.NewResId(svc, "svc"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "svc",
				},
				"spec": map[string]interface{}{
					"type": "LoadBalancer",
				},
			}),
	}
	l := makeLoader2(t)
	app, err := New(l)
	if err != nil {
		t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := app.RawResources()
	if err != nil {
		t.Fatalf("Unexpected RawResources error %v", err)
	}

	if err := expected.ErrorIfNotEqual(actual); err != nil {
		t.Fatalf("unexpected inequality: %v", err)
	}
}
