// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package namespace_test

import (
	"bytes"
	"log"
	"os"

	"sigs.k8s.io/kustomize/api/filters/namespace"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func ExampleFilter() {
	fss := builtinconfig.MakeDefaultConfig().NameSpace
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
  namespace: bar
`)}},
		Filters: []kio.Filter{namespace.Filter{Namespace: "app", FsSlice: fss}},
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
	//   namespace: app
	// ---
	// apiVersion: example.com/v1
	// kind: Bar
	// metadata:
	//   name: instance
	//   namespace: app
}
