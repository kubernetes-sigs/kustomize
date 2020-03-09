// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package starlark_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/starlark"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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

func ExampleFilter_Filter_functionConfig() {
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

	fc, err := yaml.Parse(`
kind: AnnotationSetter
spec:
  value: "hello world"
`)
	if err != nil {
		log.Fatal(err)
	}

	// fltr transforms the input using a starlark program
	fltr := &starlark.Filter{
		Name: "annotate",
		Program: `
def run(items, value):
  for item in items:
    item["metadata"]["annotations"]["foo"] = value

run(resourceList["items"], resourceList["functionConfig"]["spec"]["value"])
`,
		FunctionConfig: fc,
	}

	// output contains the transformed resources
	output := &bytes.Buffer{}

	// run the fltr against the inputs using a kio.Pipeline
	err = kio.Pipeline{
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
	//     foo: hello world
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
	//     foo: hello world
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: nginx
	//         image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image-2"}
}

// ExampleFilter_Filter_file applies a starlark program in a local file to a collection of
// resource configuration read from a directory.
func ExampleFilter_Filter_file() {
	// setup the configuration
	d, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(d)

	err = ioutil.WriteFile(filepath.Join(d, "deploy1.yaml"), []byte(`
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
`), 0600)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(d, "deploy2.yaml"), []byte(`
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
`), 0600)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(d, "annotate.star"), []byte(`
def run(items):
  for item in items:
    item["metadata"]["annotations"]["foo"] = "bar"

run(resourceList["items"])
`), 0600)
	if err != nil {
		log.Fatal(err)
	}

	fltr := &starlark.Filter{
		Name: "annotate",
		Path: filepath.Join(d, "annotate.star"),
	}

	// output contains the transformed resources
	output := &bytes.Buffer{}

	// run the fltr against the inputs using a kio.Pipeline
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.LocalPackageReader{PackagePath: d}},
		Filters: []kio.Filter{fltr},
		Outputs: []kio.Writer{&kio.ByteWriter{
			Writer:           output,
			ClearAnnotations: []string{"config.kubernetes.io/path"},
		}}}.Execute()
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
