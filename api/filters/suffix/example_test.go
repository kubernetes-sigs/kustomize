// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package suffix_test

import (
	"bytes"
	"log"
	"os"

	"sigs.k8s.io/kustomize/api/filters/suffix"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func ExampleFilter() {
	err := kio.Pipeline{
		Inputs: []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
`)}},
		Filters: []kio.Filter{suffix.Filter{
			Suffix: "-baz", FieldSpec: types.FieldSpec{Path: "metadata/name"}}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: os.Stdout}},
	}.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: example.com/v1
	// kind: Foo
	// metadata:
	//   name: instance-baz
	// ---
	// apiVersion: example.com/v1
	// kind: Bar
	// metadata:
	//   name: instance-baz
}
