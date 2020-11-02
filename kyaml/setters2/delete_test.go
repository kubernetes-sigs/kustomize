// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
)

var resourcefile2 = `apiVersion: resource.dev/v1alpha1
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
          value: "2"
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "sometag"
`

func TestDelete_Filter2(t *testing.T) {
	path := filepath.Join(os.TempDir(), "resourcefile2")

	// write initial resourcefile to temp path
	err := ioutil.WriteFile(path, []byte(resourcefile2), 0666)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// add a deleter definition
	dd := DeleterDefinition{
		Name:             "image",
		DefinitionPrefix: fieldmeta.SetterDefinitionPrefix,
	}

	err = dd.DeleteFromFile(path)
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
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: "sometag"
`
	assert.Equal(t, expected, string(b))
}
