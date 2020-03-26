// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package annotations

import (
	"bytes"
	"log"
	"os"

	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func ExampleFilter() {
	fss := builtinconfig.MakeDefaultConfig().CommonAnnotations
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
		Filters: []kio.Filter{Filter{
			Annotations: map[string]string{
				"foo": "bar",
			},
			FsSlice: fss,
		}},
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
	//   annotations:
	//     foo: bar
	// ---
	// apiVersion: example.com/v1
	// kind: Bar
	// metadata:
	//   name: instance
	//   annotations:
	//     foo: bar
}
