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

package crds

import (
	"reflect"
	"sort"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/internal/loadertest"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	crdContent = `
{
	"github.com/example/pkg/apis/jingfang/v1beta1.Bee": {
		"Schema": {
			"description": "Bee",
			"properties": {
				"apiVersion": {
					"description": "APIVersion defines the versioned schema of this representation of an object. Servers should convert 
recognized schemas to the latest internal value, and may reject unrecognized values. 
More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
					"type": "string"
				},
				"kind": {
					"description": "Kind is a string value representing the REST resource this object represents. Servers may infer
this from the endpoint the client submits requests to. Cannot be updated. In CamelCase.
More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
					"type": "string"
				},
				"metadata": {
					"$ref": "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"
				},
				"spec": {
					"$ref": "github.com/example/pkg/apis/jingfang/v1beta1.BeeSpec"
				},
				"status": {
					"$ref": "github.com/example/pkg/apis/jingfang/v1beta1.BeeStatus"
				}
			}
		},
		"Dependencies": [
			"github.com/example/pkg/apis/jingfang/v1beta1.BeeSpec",
			"github.com/example/pkg/apis/jingfang/v1beta1.BeeStatus",
			"k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"
		]
	},
	"github.com/example/pkg/apis/jingfang/v1beta1.BeeSpec": {
		"Schema": {
			"description": "BeeSpec defines the desired state of Bee"
		},
		"Dependencies": []
	},
	"github.com/example/pkg/apis/jingfang/v1beta1.BeeStatus": {
		"Schema": {
			"description": "BeeStatus defines the observed state of Bee"
		},
		"Dependencies": []
	},
	"github.com/example/pkg/apis/jingfang/v1beta1.MyKind": {
		"Schema": {
			"description": "MyKind",
			"properties": {
				"apiVersion": {
					"description": "APIVersion defines the versioned schema of this representation of an object.
Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values.
More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
					"type": "string"
				},
				"kind": {
					"description": "Kind is a string value representing the REST resource this object represents.
Servers may infer this from the endpoint the client submits requests to. Cannot be updated.
In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
					"type": "string"
				},
				"metadata": {
					"$ref": "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"
				},
				"spec": {
					"$ref": "github.com/example/pkg/apis/jingfang/v1beta1.MyKindSpec"
				},
				"status": {
					"$ref": "github.com/example/pkg/apis/jingfang/v1beta1.MyKindStatus"
				}
			}
		},
		"Dependencies": [
			"github.com/example/pkg/apis/jingfang/v1beta1.MyKindSpec",
			"github.com/example/pkg/apis/jingfang/v1beta1.MyKindStatus",
			"k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"
		]
	},
	"github.com/example/pkg/apis/jingfang/v1beta1.MyKindSpec": {
		"Schema": {
			"description": "MyKindSpec defines the desired state of MyKind",
			"properties": {
				"beeRef": {
					"x-kubernetes-object-ref-api-version": "v1beta1",
					"x-kubernetes-object-ref-kind": "Bee",
					"$ref": "github.com/example/pkg/apis/jingfang/v1beta1.Bee"
				},
				"secretRef": {
					"description": "If defined, we use this secret for configuring the MYSQL_ROOT_PASSWORD 
If it is not set we generate a secret dynamically",
					"x-kubernetes-object-ref-api-version": "v1",
					"x-kubernetes-object-ref-kind": "Secret",
					"$ref": "k8s.io/api/core/v1.LocalObjectReference"
				}
			}
		},
		"Dependencies": [
			"github.com/example/pkg/apis/jingfang/v1beta1.Bee",
			"k8s.io/api/core/v1.LocalObjectReference"
		]
	},
	"github.com/example/pkg/apis/jingfang/v1beta1.MyKindStatus": {
		"Schema": {
			"description": "MyKindStatus defines the observed state of MyKind"
		},
		"Dependencies": []
	}
}
`
)

func makeLoader(t *testing.T) loader.Loader {
	ldr := loadertest.NewFakeLoader("/testpath")
	err := ldr.AddFile("/testpath/crd.json", []byte(crdContent))
	if err != nil {
		t.Fatalf("Failed to setup fake ldr.")
	}
	return ldr
}

func TestRegisterCRD(t *testing.T) {
	refpathconfigs := []transformers.ReferencePathConfig{
		transformers.NewReferencePathConfig(
			schema.GroupVersionKind{Kind: "Bee", Version: "v1beta1"},
			[]transformers.PathConfig{
				{
					CreateIfNotPresent: false,
					GroupVersionKind:   &schema.GroupVersionKind{Kind: "MyKind"},
					Path:               []string{"spec", "beeRef", "name"},
				},
			},
		),
		transformers.NewReferencePathConfig(
			schema.GroupVersionKind{Kind: "Secret", Version: "v1"},
			[]transformers.PathConfig{
				{
					CreateIfNotPresent: false,
					GroupVersionKind:   &schema.GroupVersionKind{Kind: "MyKind"},
					Path:               []string{"spec", "secretRef", "name"},
				},
			},
		),
	}

	sort.Slice(refpathconfigs, func(i, j int) bool {
		return refpathconfigs[i].GVK() < refpathconfigs[j].GVK()
	})

	expected := []pathConfigs{
		{
			namereferencePathConfigs: refpathconfigs,
		},
	}

	ldr := makeLoader(t)

	pathconfig, _ := registerCRD(ldr, "/testpath/crd.json")

	sort.Slice(pathconfig[0].namereferencePathConfigs, func(i, j int) bool {
		return pathconfig[0].namereferencePathConfigs[i].GVK() < pathconfig[0].namereferencePathConfigs[j].GVK()
	})

	if !reflect.DeepEqual(pathconfig, expected) {
		t.Fatalf("expected\n %v\n but got\n %v\n", expected, pathconfig)
	}
}
