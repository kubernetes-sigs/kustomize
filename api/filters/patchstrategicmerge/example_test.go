// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patchstrategicmerge

import (
	"bytes"
	"log"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func ExampleFilter() {
	err := kio.Pipeline{
		Inputs: []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  replicas: 3
`)}},
		Filters: []kio.Filter{Filter{
			Patch: yaml.MustParse(`
spec:
  template:
    containers:
    - image: nginx
`),
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
	// spec:
	//   replicas: 3
	//   template:
	//     containers:
	//     - image: nginx
}
