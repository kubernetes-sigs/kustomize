// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestAddSchema(t *testing.T) {
	// reset package vars
	globalSchema = openapiData{}

	err := AddSchema(additionalSchema)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	s, err := GetSchema(`{"$ref": "#/definitions/io.k8s.config.setters.replicas"}`, Schema())
	if !assert.GreaterOrEqual(t, len(globalSchema.schema.Definitions), 1) {
		t.FailNow()
	}
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, `map[x-kustomize:map[setBy:Jane setter:map[name:replicas value:5]]]`,
		fmt.Sprintf("%v", s.Schema.Extensions))
}

var additionalSchema = []byte(`
{
  "definitions": {
    "io.k8s.config.setters.replicas": {
      "description": "replicas description.",
      "type": "integer",
      "x-kustomize": {"setBy":"Jane","setter": {"name":"replicas","value":"5"}}
    }
  },
  "invalid": "field"
}
`)

func TestSchemaForResourceType(t *testing.T) {
	ResetOpenAPI()
	defer ResetOpenAPI()

	// The builtin schema document is no longer embedded: the full-schema
	// lookup has nothing to serve for builtin types...
	s := SchemaForResourceType(
		yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	if !assert.Nil(t, s) {
		t.FailNow()
	}

	// ...but strategic-merge metadata is served from the precomputed table.
	pm := PatchMetaSchemaForResourceType(
		yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	if !assert.NotNil(t, pm) {
		t.FailNow()
	}
	containers := pm.Lookup("spec", "template", "spec", "containers")
	if !assert.NotNil(t, containers) {
		t.FailNow()
	}
	strategy, key := containers.PatchStrategyAndKey()
	assert.Equal(t, "merge", strategy)
	assert.Equal(t, "name", key)

	// A supplied schema document restores the full contract, descriptions
	// included, and overrides the table for the types it defines.
	if !assert.NoError(t, AddSchema(deploymentSchema)) {
		t.FailNow()
	}
	s = SchemaForResourceType(
		yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	if !assert.NotNil(t, s) {
		t.FailNow()
	}
	f := s.Field("spec")
	if !assert.NotNil(t, f) {
		t.FailNow()
	}
	assert.Equal(t, "The desired behavior of the Deployment.", f.Schema.Description)

	pm = PatchMetaSchemaForResourceType(
		yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	if !assert.NotNil(t, pm) {
		t.FailNow()
	}
	// the supplied document wins over the table for this type
	assert.Equal(t, "The desired behavior of the Deployment.",
		pm.Field("spec").Schema.Description)
}

var deploymentSchema = []byte(`
{
  "definitions": {
    "io.k8s.api.apps.v1.Deployment": {
      "x-kubernetes-group-version-kind": [{"group": "apps", "kind": "Deployment", "version": "v1"}],
      "properties": {
        "spec": {
          "description": "The desired behavior of the Deployment.",
          "type": "object"
        }
      }
    }
  }
}
`)

func TestSchemaFromFile(t *testing.T) {
	ResetOpenAPI()
	inputyaml := `
openAPI:
  definitions:
    io.k8s.cli.setters.image-name:
      x-k8s-cli:
        setter:
          name: image-name
          value: "nginx"
 `
	f, err := os.CreateTemp("", "openapi-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, os.WriteFile(f.Name(), []byte(inputyaml), 0o600)) {
		t.FailNow()
	}

	sc, err := SchemaFromFile(f.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	s, err := GetSchema(`{"$ref": "#/definitions/io.k8s.cli.setters.image-name"}`, sc)

	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Greater(t, len(sc.Definitions), 0) {
		t.FailNow()
	}
	assert.Equal(t, `map[x-k8s-cli:map[setter:map[name:image-name value:nginx]]]`,
		fmt.Sprintf("%v", s.Schema.Extensions))
}

func TestPopulateDefsInOpenAPI_Substitution(t *testing.T) {
	ResetOpenAPI()
	inputyaml := `
openAPI:
  definitions:
    io.k8s.cli.setters.image-name:
      x-k8s-cli:
        setter:
          name: image-name
          value: "nginx"
    io.k8s.cli.setters.image-tag:
      x-k8s-cli:
        setter:
          name: image-tag
          value: "1.8.1"
    io.k8s.cli.substitutions.image:
      x-k8s-cli:
        substitution:
          name: image
          pattern: IMAGE_NAME:IMAGE_TAG
          values:
          - marker: "IMAGE_NAME"
            ref: "#/definitions/io.k8s.cli.setters.image-name"
          - marker: "IMAGE_TAG"
            ref: "#/definitions/io.k8s.cli.setters.image-tag"
 `

	f, err := os.CreateTemp("", "openapi-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, os.WriteFile(f.Name(), []byte(inputyaml), 0o600)) {
		t.FailNow()
	}

	sc, err := SchemaFromFile(f.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	s, err := GetSchema(`{"$ref": "#/definitions/io.k8s.cli.substitutions.image"}`, sc)

	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, len(sc.Definitions), 3) {
		t.FailNow()
	}

	assert.Equal(t,
		`map[x-k8s-cli:map[substitution:map[name:image pattern:IMAGE_NAME:IMAGE_TAG`+
			` values:[map[marker:IMAGE_NAME ref:#/definitions/io.k8s.cli.setters.image-name]`+
			` map[marker:IMAGE_TAG ref:#/definitions/io.k8s.cli.setters.image-tag]]]]]`,
		fmt.Sprintf("%v", s.Schema.Extensions))
}

func TestAddSchemaFromFile_empty(t *testing.T) {
	ResetOpenAPI()
	inputyaml := `
kind: Example
 `

	f, err := os.CreateTemp("", "openapi-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, os.WriteFile(f.Name(), []byte(inputyaml), 0o600)) {
		t.FailNow()
	}

	sc, err := SchemaFromFile(f.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Nil(t, sc) {
		t.FailNow()
	}
}

func TestIsNamespaceScoped_builtin(t *testing.T) {
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
			ResetOpenAPI()
			isNamespaceable, isFound := IsNamespaceScoped(test.typeMeta)

			if !test.expectIsFound {
				assert.False(t, isFound)
				return
			}
			assert.True(t, isFound)
			assert.Equal(t, test.expectIsNamespaced, isNamespaceable)
		})
	}
}

// TestIsNamespaceScopedPrecompute checks that precomputedIsNamespaceScoped includes
// every resource type from the built-in OpenAPI schema with the correct namespace scope.
func TestIsNamespaceScopedPrecompute(t *testing.T) {
	initSchema()
	for k, actual := range globalSchema.namespaceabilityByResourceType {
		expected, ok := precomputedIsNamespaceScoped[k]
		if !ok {
			t.Fatalf("resource type %v found in globalSchema but missing from precomputedIsNamespaceScoped", k)
		}
		if actual != expected {
			t.Fatalf("namespaceability mismatch for %v: expected %v got %v", k, expected, actual)
		}
	}
}

func TestIsNamespaceScoped_custom(t *testing.T) {
	SuppressBuiltInSchemaUse()
	err := AddSchema([]byte(`
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
        }
      }
    },
    "/apis/custom.io/v1/clustercustoms": {
      "get": {
        "x-kubernetes-action": "get",
        "x-kubernetes-group-version-kind": {
          "group": "custom.io",
          "kind": "ClusterCustom",
          "version": "v1"
        }
      }
    }
  }
}
`))
	require.NoError(t, err)

	isNamespaceable, isFound := IsNamespaceScoped(yaml.TypeMeta{
		APIVersion: "custom.io/v1",
		Kind:       "ClusterCustom",
	})
	assert.True(t, isFound)
	assert.False(t, isNamespaceable)

	isNamespaceable, isFound = IsNamespaceScoped(yaml.TypeMeta{
		APIVersion: "custom.io/v1",
		Kind:       "Custom",
	})
	assert.True(t, isFound)
	assert.True(t, isNamespaceable)
}

func TestCanSetAndResetSchemaConcurrently(t *testing.T) {
	t.Run("SetSchema doesn't cause a data race when called concurrently", func(t *testing.T) {
		set := func(wg *sync.WaitGroup) {
			defer wg.Done()
			err := SetSchema(
				map[string]string{
					"/apis/custom.io/v1": "true",
				},
				[]byte(`
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
						}
					  }
					},
					"/apis/custom.io/v1/clustercustoms": {
					  "get": {
						"x-kubernetes-action": "get",
						"x-kubernetes-group-version-kind": {
						  "group": "custom.io",
						  "kind": "ClusterCustom",
						  "version": "v1"
						}
					  }
					}
				  }
				}
			`),
				true,
			)
			require.NoError(t, err)
		}

		var wg sync.WaitGroup
		require.NotPanics(t, func() {
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go set(&wg)
			}
		})
		wg.Wait()
	})

	t.Run("ResetOpenAPI doesn't cause a data race when called concurrently", func(t *testing.T) {
		reset := func(wg *sync.WaitGroup) {
			defer wg.Done()
			ResetOpenAPI()
		}
		var wg sync.WaitGroup
		require.NotPanics(t, func() {
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go reset(&wg)
			}
		})
		wg.Wait()
	})
}
