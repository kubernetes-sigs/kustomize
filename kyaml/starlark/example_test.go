// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package starlark_test

import (
	"bytes"
	"fmt"
	"log"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/starlark"
)

func ExampleFilter_Filter() {
	// input contains the items that will provided to the starlark program
	input := bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-1
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image-1"}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-2
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image-2"}
`)

	// fltr transforms the input using a starlark program
	fltr := &starlark.Filter{
		Name: "annotate",
		Program: `
def run(items):
  for item in items:
    item["metadata"]["annotations"]["foo"] = "bar"

run(resourceList["items"])
`,
	}

	// output contains the transformed resources
	output := &bytes.Buffer{}

	// run the fltr against the inputs using a kio.Pipeline
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: input}},
		Filters: []kio.Filter{fltr},
		Outputs: []kio.Writer{&kio.ByteWriter{Writer: output}}}.Execute()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(output.String())

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: deployment-1
	//   annotations:
	//     foo: bar
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: nginx
	//         image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image-1"}
	//---
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: deployment-2
	//   annotations:
	//     foo: bar
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: nginx
	//         image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image-2"}
}
