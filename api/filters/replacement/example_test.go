// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package replacement

import (
	"bytes"
	"log"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func ExampleFilter() {
	f := Filter{}
	err := yaml.Unmarshal([]byte(`
replacements:
- source:
    kind: Foo2
    fieldPath: spec.replicas
  targets:
  - select:
      kind: Foo1
    fieldPaths: 
    - spec.replicas`), &f)
	if err != nil {
		log.Fatal(err)
	}

	err = kio.Pipeline{
		Inputs: []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: example.com/v1
kind: Foo1
metadata:
  name: instance
spec:
  replicas: 3
---
apiVersion: example.com/v1
kind: Foo2
metadata:
  name: instance
spec:
  replicas: 99
`)}},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: os.Stdout}},
	}.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: example.com/v1
	// kind: Foo1
	// metadata:
	//   name: instance
	// spec:
	//   replicas: 99
	// ---
	// apiVersion: example.com/v1
	// kind: Foo2
	// metadata:
	//   name: instance
	// spec:
	//   replicas: 99
}
