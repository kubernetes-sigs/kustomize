// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAllSetterDefinitions(t *testing.T) {
	srcOpenAPIFile := `openAPI:
  definitions:
    io.k8s.cli.setters.namespace:
      x-k8s-cli:
        setter:
          name: namespace
          value: "project-namespace"
    io.k8s.cli.setters.replicas:
      x-k8s-cli:
        setter:
          name: replicas
          value: "4"`

	destFile1 := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: some-other-namespace # {"$ref": "#/definitions/io.k8s.cli.setters.namespace"}
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}`

	destFile2 := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: some-other-namespace2 # {"$ref": "#/definitions/io.k8s.cli.setters.namespace"}
spec:
  replicas: 2 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}`

	expectedDestFile := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: project-namespace # {"$ref": "#/definitions/io.k8s.cli.setters.namespace"}
spec:
  replicas: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}`

	srcDir, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	destDir1, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	destDir2, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	defer os.RemoveAll(srcDir)
	defer os.RemoveAll(destDir1)
	defer os.RemoveAll(destDir2)

	err = ioutil.WriteFile(srcDir+"/OpenAPIFile", []byte(srcOpenAPIFile), 0600)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = ioutil.WriteFile(destDir1+"/destFile.yaml", []byte(destFile1), 0600)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = ioutil.WriteFile(destDir2+"/destFile.yaml", []byte(destFile2), 0600)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = SetAllSetterDefinitions(srcDir+"/OpenAPIFile", destDir1, destDir2)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	actualdestFile1, err := ioutil.ReadFile(destDir1 + "/destFile.yaml")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, strings.Trim(string(actualdestFile1), "\n"), expectedDestFile) {
		t.FailNow()
	}

	actualdestFile2, err := ioutil.ReadFile(destDir2 + "/destFile.yaml")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, strings.Trim(string(actualdestFile2), "\n"), expectedDestFile) {
		t.FailNow()
	}
}
