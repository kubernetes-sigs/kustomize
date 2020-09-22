// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAllSetterDefinitions(t *testing.T) {
	var tests = []struct {
		name                    string
		srcOpenAPIFile          string
		destFile                string
		destOpenAPI             string
		expectedDestFile        string
		expectedDestOpenAPIFile string
		syncSchema              bool
	}{
		{
			name:       "set definitions with syncSchema",
			syncSchema: true,
			srcOpenAPIFile: `openAPI:
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
          value: "4"`,

			destFile: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: some-other-namespace # {"$ref": "#/definitions/io.k8s.cli.setters.namespace"}
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas-new"}`,

			destOpenAPI: `openAPI:
  definitions:
    io.k8s.cli.setters.namespace:
      x-k8s-cli:
        setter:
          name: namespace
          value: "some-other-namespace"
    io.k8s.cli.setters.replicas-new:
      x-k8s-cli:
        setter:
          name: replicas-new
          value: "3"`,

			expectedDestFile: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: project-namespace # {"$ref": "#/definitions/io.k8s.cli.setters.namespace"}
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas-new"}
`,

			expectedDestOpenAPIFile: `openAPI:
  definitions:
    io.k8s.cli.setters.namespace:
      x-k8s-cli:
        setter:
          name: namespace
          value: "project-namespace"
          isSet: true
    io.k8s.cli.setters.replicas-new:
      x-k8s-cli:
        setter:
          name: replicas-new
          value: "3"
`,
		},
		{
			name:       "set values only to resources and not the openAPI",
			syncSchema: false,
			srcOpenAPIFile: `openAPI:
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
          value: "4"`,

			destFile: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: some-other-namespace # {"$ref": "#/definitions/io.k8s.cli.setters.namespace"}
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas-new"}`,

			destOpenAPI: `openAPI:
  definitions:
    io.k8s.cli.setters.namespace:
      x-k8s-cli:
        setter:
          name: namespace
          value: "some-other-namespace"
    io.k8s.cli.setters.replicas-new:
      x-k8s-cli:
        setter:
          name: replicas-new
          value: "3"`,

			expectedDestFile: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: project-namespace # {"$ref": "#/definitions/io.k8s.cli.setters.namespace"}
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas-new"}
`,

			expectedDestOpenAPIFile: `openAPI:
  definitions:
    io.k8s.cli.setters.namespace:
      x-k8s-cli:
        setter:
          name: namespace
          value: "some-other-namespace"
    io.k8s.cli.setters.replicas-new:
      x-k8s-cli:
        setter:
          name: replicas-new
          value: "3"`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			srcDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			destDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			defer os.RemoveAll(srcDir)
			defer os.RemoveAll(destDir)

			err = ioutil.WriteFile(filepath.Join(srcDir, "Krmfile"), []byte(test.srcOpenAPIFile), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			err = ioutil.WriteFile(filepath.Join(destDir, "destFile.yaml"), []byte(test.destFile), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			err = ioutil.WriteFile(filepath.Join(destDir, "Krmfile"), []byte(test.destOpenAPI), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			err = SetAllSetterDefinitions(test.syncSchema, filepath.Join(srcDir, "Krmfile"), destDir)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actualDestFile1, err := ioutil.ReadFile(filepath.Join(destDir, "destFile.yaml"))
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t, test.expectedDestFile, string(actualDestFile1)) {
				t.FailNow()
			}

			actualDestOpenAPIFile1, err := ioutil.ReadFile(filepath.Join(destDir, "Krmfile"))
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t, test.expectedDestOpenAPIFile, string(actualDestOpenAPIFile1)) {
				t.FailNow()
			}
		})
	}
}
