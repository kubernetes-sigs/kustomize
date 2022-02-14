// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapignostic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestAddOpenApiDefinitions(t *testing.T) {
	parseBuiltin("v1212")
	// The length here may change when the builtin openapi schema is updated.
	assert.Equal(t, len(globalOpenApiSchema.schema.Definitions.AdditionalProperties), 623)
	namedSchema0 := globalOpenApiSchema.schema.Definitions.AdditionalProperties[0]
	assert.Equal(t, namedSchema0.GetName(), "io.k8s.api.admissionregistration.v1.MutatingWebhook")
}

func TestSchemaForResourceType(t *testing.T) {
	ResetSchema()
	parseBuiltin("v1212")

	s := SchemaForResourceType(
		yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	if !assert.NotNil(t, s) {
		t.FailNow()
	}
	assert.Equal(t, s.GetName(), "io.k8s.api.apps.v1.Deployment")
	assert.Contains(t, s.GetValue().String(), "description:\"Deployment enables declarative updates for Pods and ReplicaSets.\"")
}


func TestFindNamespaceability_builtin(t *testing.T) {
	testCases := []struct {
		name               string
		typeMeta           yaml.TypeMeta
		expectIsFound      bool
		expectIsNamespaced bool
	}{
		{
			name: "namespacescoped resource",
			typeMeta: yaml.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			expectIsFound:      true,
			expectIsNamespaced: true,
		},
		{
			name: "clusterscoped resource",
			typeMeta: yaml.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			expectIsFound:      true,
			expectIsNamespaced: false,
		},
		{
			name: "unknown resource",
			typeMeta: yaml.TypeMeta{
				APIVersion: "custom.io/v1",
				Kind:       "Custom",
			},
			expectIsFound: false,
		},
	}

	for i := range testCases {
		test := testCases[i]
		t.Run(test.name, func(t *testing.T) {
			ResetSchema()
			parseBuiltin("v1212")

			isNamespaceable, isFound := namespaceabilityFromSchema(test.typeMeta)

			if !test.expectIsFound {
				assert.False(t, isFound)
				return
			}
			assert.True(t, isFound)
			assert.Equal(t, test.expectIsNamespaced, isNamespaceable)
		})
	}
}

func TestFindNamespaceability_custom(t *testing.T) {
	ResetSchema()
	err := parseDocument([]byte(`
{
 "definitions": {},
 "paths": {
   "/apis/custom.io/v1/namespaces/{namespace}/customs/{name}": {
     "get": {
       "x-kubernetes-action": "get",
       "x-kubernetes-group-version-kind": {
         "group": "custom.io",
         "kind": "Custom",
         "version": "v1"
       },
       "responses": []
     }
   },
   "/apis/custom.io/v1/clustercustoms": {
     "get": {
       "x-kubernetes-action": "get",
       "x-kubernetes-group-version-kind": {
         "group": "custom.io",
         "kind": "ClusterCustom",
         "version": "v1"
       },
       "responses": []
     }
   }
 },
 "info": {
   "title": "Kustomization",
   "version": "v1beta1"
 },
 "swagger": "2.0",
 "paths": {}
}
`))
	assert.NoError(t, err)

	isNamespaceable, isFound := namespaceabilityFromSchema(yaml.TypeMeta{
		APIVersion: "custom.io/v1",
		Kind:       "ClusterCustom",
	})
	assert.True(t, isFound)
	assert.False(t, isNamespaceable)

	isNamespaceable, isFound = namespaceabilityFromSchema(yaml.TypeMeta{
		APIVersion: "custom.io/v1",
		Kind:       "Custom",
	})
	assert.True(t, isFound)
	assert.True(t, isNamespaceable)
}

func namespaceabilityFromSchema(typeMeta yaml.TypeMeta) (bool, bool) {
	isNamespaceScoped, found := globalOpenApiSchema.namespaceabilityByResourceType[typeMeta]
	return isNamespaceScoped, found
}
