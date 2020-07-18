// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec_test

import (
	"bytes"
	"log"
	"os"

	. "sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func ExampleFilter() {
	in := &kio.ByteReader{
		Reader: bytes.NewBufferString(`
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
`),
	}
	fltr := Filter{
		CreateKind: yaml.ScalarNode,
		SetValue:   filtersutil.SetScalar("green"),
		FieldSpec:  types.FieldSpec{Path: "a/b", CreateIfNotPresent: true},
	}

	err := kio.Pipeline{
		Inputs:  []kio.Reader{in},
		Filters: []kio.Filter{kio.FilterAll(fltr)},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: os.Stdout}},
	}.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: example.com/v1
	// kind: Foo
	// metadata:
	//   name: instance
	// a:
	//   b: green
	// ---
	// apiVersion: example.com/v1
	// kind: Bar
	// metadata:
	//   name: instance
	// a:
	//   b: green
}
