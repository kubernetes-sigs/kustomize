// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddUpdateSubstitution(t *testing.T) {
	path := os.TempDir() + "/resourcefile"

	//write initial resourcefile to temp path
	err := ioutil.WriteFile(path, []byte(resourcefile), 0666)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	value1 := Value{
		Marker: "IMAGE_NAME",
		Ref:    "#/definitions/io.k8s.cli.setters.image-name",
	}

	value2 := Value{
		Marker: "IMAGE_TAG",
		Ref:    "#/definitions/io.k8s.cli.setters.image-tag",
	}

	values := []Value{value1, value2}

	//add a setter definition
	subd := SubstitutionDefinition{
		Name:    "image",
		Pattern: "IMAGE_NAME:IMAGE_TAG",
		Values:  values,
	}

	err = subd.AddSubstitutionToFile(path)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// update setter definition
	subd2 := SubstitutionDefinition{
		Name:    "image",
		Pattern: "IMAGE_NAME:IMAGE_TAG2",
	}

	err = subd2.AddSubstitutionToFile(path)

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
        substitution:
          name: image
          pattern: IMAGE_NAME:IMAGE_TAG2
`
	assert.Equal(t, expected, string(b))
}
