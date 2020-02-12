// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var resourcefile = `apiVersion: resource.dev/v1alpha1
kind: resourcefile
metadata:
    name: hello-world-set
upstream:
    type: git
    git:
        commit: 5c1c019b59299a4f6c7edd1ff5ff54d720621bbe
        directory: /package-examples/helloworld-set
        ref: v0.1.0
packageMetadata:
    shortDescription: example package using setters`

func TestAddUpdateSetter(t *testing.T) {
	path := os.TempDir() + "/resourcefile"

	//write initial resourcefile to temp path
	err := ioutil.WriteFile(path, []byte(resourcefile), 0666)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	//add a setter definition
	sd := SetterDefinition{
		Name:  "image",
		Value: "1",
	}

	err = sd.AddSetterToFile(path)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// update setter definition
	sd2 := SetterDefinition{
		Name:  "image",
		Value: "2",
	}

	err = sd2.AddSetterToFile(path)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.FailNow()
	}

	expected := `apiVersion: resource.dev/v1alpha1
kind: resourcefile
metadata:
  name: hello-world-set
upstream:
  type: git
  git:
    commit: 5c1c019b59299a4f6c7edd1ff5ff54d720621bbe
    directory: /package-examples/helloworld-set
    ref: v0.1.0
packageMetadata:
  shortDescription: example package using setters
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: 2
`
	assert.Equal(t, expected, string(b))
}
