// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func TestCheckRequiredSettersSet(t *testing.T) {
	var tests = []struct {
		name             string
		inputOpenAPIfile string
		expectedError    bool
	}{
		{
			name: "no required, no isSet",
			inputOpenAPIfile: `
apiVersion: v1alpha1
kind: OpenAPIfile
openAPI:
  definitions:
    io.k8s.cli.setters.gcloud.project.projectNumber:
      description: hello world
      x-k8s-cli:
        setter:
          name: gcloud.project.projectNumber
          value: "123"
          setBy: me
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
 `,
			expectedError: false,
		},
		{
			name: "required true, no isSet",
			inputOpenAPIfile: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
          required: true
 `,
			expectedError: true,
		},
		{
			name: "required true, isSet true",
			inputOpenAPIfile: `
apiVersion: v1alpha1
kind: Example
openAPI:
  definitions:
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
          required: true
          isSet: true
 `,
			expectedError: false,
		},

		{
			name: "required false, isSet true",
			inputOpenAPIfile: `
apiVersion: v1alpha1
kind: OpenAPIfile
openAPI:
  definitions:
    io.k8s.cli.setters.gcloud.project.projectNumber:
      description: hello world
      x-k8s-cli:
        setter:
          name: gcloud.project.projectNumber
          value: "123"
          setBy: me
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
          required: false
          isSet: true
 `,
			expectedError: false,
		},

		{
			name: "required true, isSet false",
			inputOpenAPIfile: `
apiVersion: v1alpha1
kind: OpenAPIfile
openAPI:
  definitions:
    io.k8s.cli.setters.gcloud.project.projectNumber:
      description: hello world
      x-k8s-cli:
        setter:
          name: gcloud.project.projectNumber
          value: "123"
          setBy: me
    io.k8s.cli.setters.replicas:
      description: hello world
      x-k8s-cli:
        setter:
          name: replicas
          value: "3"
          setBy: me
          required: true
          isSet: false
 `,
			expectedError: true,
		},

		{
			name:             "no openAPI",
			inputOpenAPIfile: ``,
			expectedError:    false,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			openapi.ResetOpenAPI()
			defer openapi.ResetOpenAPI()
			dir, err := ioutil.TempDir("", "")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)
			err = ioutil.WriteFile(filepath.Join(dir, "Krmfile"), []byte(test.inputOpenAPIfile), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			clean, err := openapi.AddSchemaFromFile(filepath.Join(dir, "Krmfile"))
			defer clean()
			if err != nil {
				// do nothing if openAPI file or schema doesn't exist, CheckRequiredSettersSet()
				// should not throw any error
				fmt.Println("Unable to load schema from file, continuing...")
			}
			err = CheckRequiredSettersSet()
			if test.expectedError && !assert.Error(t, err) {
				t.FailNow()
			}
			if !test.expectedError && !assert.NoError(t, err) {
				t.FailNow()
			}
		})
	}
}
