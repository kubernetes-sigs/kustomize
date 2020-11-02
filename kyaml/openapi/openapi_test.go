// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestAddSchema(t *testing.T) {
	// reset package vars
	globalSchema = openapiData{}

	err := AddSchema(additionalSchema)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	s, err := GetSchema(`{"$ref": "#/definitions/io.k8s.config.setters.replicas"}`)
	if !assert.Greater(t, len(globalSchema.schema.Definitions), 200) {
		t.FailNow()
	}
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, `map[x-kustomize:map[setBy:Jane setter:map[name:replicas value:5]]]`,
		fmt.Sprintf("%v", s.Schema.Extensions))
}

func TestNoUseBuiltInSchema_AddSchema(t *testing.T) {
	// reset package vars
	globalSchema = openapiData{}

	SuppressBuiltInSchemaUse()
	err := AddSchema(additionalSchema)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	s, err := GetSchema(`{"$ref": "#/definitions/io.k8s.config.setters.replicas"}`)
	if !assert.Equal(t, len(globalSchema.schema.Definitions), 1) {
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
	// reset package vars
	globalSchema = openapiData{}

	s := SchemaForResourceType(
		yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	if !assert.NotNil(t, s) {
		t.FailNow()
	}

	f := s.Field("spec")
	if !assert.NotNil(t, f) {
		t.FailNow()
	}
	if !assert.Equal(t, "DeploymentSpec is the specification of the desired behavior of the Deployment.",
		f.Schema.Description) {
		t.FailNow()
	}

	replicas := f.Field("replicas")
	if !assert.NotNil(t, replicas) {
		t.FailNow()
	}
	if !assert.Equal(t, "Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1.",
		replicas.Schema.Description) {
		t.FailNow()
	}

	temp := f.Field("template")
	if !assert.NotNil(t, temp) {
		t.FailNow()
	}
	if !assert.Equal(t, "PodTemplateSpec describes the data a pod should have when created from a template",
		temp.Schema.Description) {
		t.FailNow()
	}

	containers := temp.Field("spec").Field("containers").Elements()
	if !assert.NotNil(t, containers) {
		t.FailNow()
	}

	targetPort := containers.Field("ports").Elements().Field("containerPort")
	if !assert.NotNil(t, targetPort) {
		t.FailNow()
	}
	if !assert.Equal(t, "Number of port to expose on the pod's IP address. This must be a valid port number, 0 < x < 65536.",
		targetPort.Schema.Description) {
		t.FailNow()
	}

	arg := containers.Field("args").Elements()
	if !assert.NotNil(t, arg) {
		t.FailNow()
	}
	if !assert.Equal(t, "string", arg.Schema.Type[0]) {
		t.FailNow()
	}
}

func TestAddSchemaFromFile(t *testing.T) {
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
	f, err := ioutil.TempFile("", "openapi-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, ioutil.WriteFile(f.Name(), []byte(inputyaml), 0600)) {
		t.FailNow()
	}

	clean, err := AddSchemaFromFile(f.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer clean()

	s, err := GetSchema(`{"$ref": "#/definitions/io.k8s.cli.setters.image-name"}`)

	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Greater(t, len(globalSchema.schema.Definitions), 200) {
		t.FailNow()
	}
	assert.Equal(t, `map[x-k8s-cli:map[setter:map[name:image-name value:nginx]]]`,
		fmt.Sprintf("%v", s.Schema.Extensions))
}

func TestDeleteSchemaInFile(t *testing.T) {
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
	f, err := ioutil.TempFile("", "openapi-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, ioutil.WriteFile(f.Name(), []byte(inputyaml), 0600)) {
		t.FailNow()
	}

	clean, err := AddSchemaFromFile(f.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	s, err := GetSchema(`{"$ref": "#/definitions/io.k8s.cli.setters.image-name"}`)

	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Greater(t, len(globalSchema.schema.Definitions), 200) {
		t.FailNow()
	}
	assert.Equal(t, `map[x-k8s-cli:map[setter:map[name:image-name value:nginx]]]`,
		fmt.Sprintf("%v", s.Schema.Extensions))

	clean()

	_, err = GetSchema(`{"$ref": "#/definitions/io.k8s.cli.setters.image-name"}`)

	if !assert.Error(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `object has no key "io.k8s.cli.setters.image-name"`, err.Error()) {
		t.FailNow()
	}
}

func TestDeleteSchemaInFileNoDefs(t *testing.T) {
	ResetOpenAPI()
	inputyaml := `
openAPI:
  definitions:
 `
	f, err := ioutil.TempFile("", "openapi-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, ioutil.WriteFile(f.Name(), []byte(inputyaml), 0600)) {
		t.FailNow()
	}

	clean, err := AddSchemaFromFile(f.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer clean()
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

	f, err := ioutil.TempFile("", "openapi-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, ioutil.WriteFile(f.Name(), []byte(inputyaml), 0600)) {
		t.FailNow()
	}

	clean, err := AddSchemaFromFile(f.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer clean()

	s, err := GetSchema(`{"$ref": "#/definitions/io.k8s.cli.substitutions.image"}`)

	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Greater(t, len(globalSchema.schema.Definitions), 200) {
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

	f, err := ioutil.TempFile("", "openapi-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, ioutil.WriteFile(f.Name(), []byte(inputyaml), 0600)) {
		t.FailNow()
	}

	clean, err := AddSchemaFromFile(f.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer clean()

	if !assert.Equal(t, len(globalSchema.schema.Definitions), 0) {
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
	assert.NoError(t, err)

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
