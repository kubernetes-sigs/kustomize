// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator_test

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/loader"

	. "sigs.k8s.io/kustomize/api/internal/accumulator"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
)

// This defines two CRD's:  Bee and MyKind.
//
// Bee is boring, it's spec has no dependencies.
//
// MyKind, however, has a spec that contains
// a Bee and a (k8s native) Secret.
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

func TestLoadCRDs(t *testing.T) {
	nbrs := []builtinconfig.NameBackReferences{
		{
			Gvk: resid.Gvk{Kind: "Secret", Version: "v1"},
			FieldSpecs: []types.FieldSpec{
				{
					CreateIfNotPresent: false,
					Gvk:                resid.Gvk{Kind: "MyKind"},
					Path:               "spec/secretRef/name",
				},
			},
		},
		{
			Gvk: resid.Gvk{Kind: "Bee", Version: "v1beta1"},
			FieldSpecs: []types.FieldSpec{
				{
					CreateIfNotPresent: false,
					Gvk:                resid.Gvk{Kind: "MyKind"},
					Path:               "spec/beeRef/name",
				},
			},
		},
	}

	expectedTc := &builtinconfig.TransformerConfig{
		NameReference: nbrs,
	}

	fSys := filesys.MakeFsInMemory()
	fSys.WriteFile("/testpath/crd.json", []byte(crdContent))
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, "/testpath", fSys)
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}

	actualTc, err := LoadConfigFromCRDs(ldr, []string{"crd.json"})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if !reflect.DeepEqual(actualTc, expectedTc) {
		t.Fatalf("expected\n %v\n but got\n %v\n", expectedTc, actualTc)
	}
}
