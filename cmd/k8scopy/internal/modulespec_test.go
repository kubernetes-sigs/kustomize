// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/cmd/k8scopy/internal"
)

var data = []byte(`module: k8s.io/apimachinery
version: v0.17.0
packages:
- name: pkg/labels
  files:
  - labels.go
  - selector.go
  - zz_generated.deepcopy.go
- name: pkg/selection
  files:
  - operator.go
- name: pkg/util/sets
  files:
  - empty.go
  - string.go
- name: pkg/util/errors
  files:
  - errors.go
- name: pkg/util/validation
  files:
  - validation.go
- name: pkg/util/validation/field
  files:
  - errors.go
  - path.go
`)

func TestReadSpec(t *testing.T) {
	fn := writeFile(t, data)
	defer os.Remove(fn)
	x := ReadSpec(fn)
	assert.Equal(t, "k8s.io/apimachinery@v0.17.0", x.Name())
	assert.Equal(t, 6, len(x.Packages))
	assert.Equal(t, "pkg/util/validation/field", x.Packages[5].Name)
	assert.Equal(t, "path.go", x.Packages[5].Files[1])
}

// Write content to temp file, returning file name.
func writeFile(t *testing.T, content []byte) string {
	f, err := ioutil.TempFile("", "testjunk")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}
