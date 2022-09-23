// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

var openAPIFile = `
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: "2"
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "sometag"
`

var resourceFile = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    image: 3 # {"$openapi":"image"}
spec:
  image: 3 # {"$openapi":"image"}
`

func TestDeleterCreator_Delete(t *testing.T) {
	openapi.ResetOpenAPI()
	defer openapi.ResetOpenAPI()
	openAPI, err := os.CreateTemp("", "openAPI.yaml")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.Remove(openAPI.Name())
	// write openapi to temp dir
	err = os.WriteFile(openAPI.Name(), []byte(openAPIFile), 0644)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// write resource file to temp dir
	resource, err := os.CreateTemp("", "k8s-cli-*.yaml")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.Remove(resource.Name())
	err = os.WriteFile(resource.Name(), []byte(resourceFile), 0644)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	sc, err := openapi.SchemaFromFile(openAPI.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	// add a delete creator
	dc := DeleterCreator{
		Name:             "image",
		DefinitionPrefix: fieldmeta.SetterDefinitionPrefix,
		SettersSchema:    sc,
	}

	dc.OpenAPIPath = openAPI.Name()
	dc.ResourcesPath = resource.Name()

	err = dc.Delete()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	actualOpenAPI, err := os.ReadFile(openAPI.Name())
	if err != nil {
		t.FailNow()
	}

	actualResource, err := os.ReadFile(resource.Name())
	if err != nil {
		t.FailNow()
	}

	expectedOpenAPI := `
openAPI:
  definitions:
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "sometag"
`
	expectedResoure := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    image: 3
spec:
  image: 3
`

	assert.Equal(t, strings.TrimSpace(expectedOpenAPI), strings.TrimSpace(string(actualOpenAPI)))
	assert.Equal(t, strings.TrimSpace(expectedResoure), strings.TrimSpace(string(actualResource)))
}
