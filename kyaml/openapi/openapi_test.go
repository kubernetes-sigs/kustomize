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

	_, err := AddSchema(additionalSchema)
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
	_, err := AddSchema(additionalSchema)
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

	err = AddSchemaFromFile(f.Name())
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

	if !assert.NoError(t, AddSchemaFromFile(f.Name())) {
		t.FailNow()
	}

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

	if !assert.NoError(t, AddSchemaFromFile(f.Name())) {
		t.FailNow()
	}

	if !assert.Equal(t, len(globalSchema.schema.Definitions), 0) {
		t.FailNow()
	}
}
