// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"io/ioutil"
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
	openAPI, err := ioutil.TempFile("", "openAPI.yaml")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.Remove(openAPI.Name())
	// write openapi to temp dir
	err = ioutil.WriteFile(openAPI.Name(), []byte(openAPIFile), 0666)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// write resource file to temp dir
	resource, err := ioutil.TempFile("", "k8s-cli-*.yaml")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.Remove(resource.Name())
	err = ioutil.WriteFile(resource.Name(), []byte(resourceFile), 0666)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// add a delete creator
	dc := DeleterCreator{
		Name:             "image",
		DefinitionPrefix: fieldmeta.SetterDefinitionPrefix,
	}

	clean, err := openapi.AddSchemaFromFile(openAPI.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer clean()
	dc.OpenAPIPath = openAPI.Name()
	dc.ResourcesPath = resource.Name()

	err = dc.Delete()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	actualOpenAPI, err := ioutil.ReadFile(openAPI.Name())
	if err != nil {
		t.FailNow()
	}

	actualResource, err := ioutil.ReadFile(resource.Name())
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
