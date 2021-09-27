// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patchjson6902

import (
	"bytes"
	"log"
	"os"

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
  namespace: bar
`)}},
		Filters: []kio.Filter{
			Filter{
				Patch: `
- op: replace
  path: /metadata/namespace
  value: "ns"
`,
			},
		},
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
	//   namespace: ns
	// ---
	// apiVersion: example.com/v1
	// kind: Bar
	// metadata:
	//   name: instance
	//   namespace: ns
}
