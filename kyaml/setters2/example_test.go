// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ExampleAdd demonstrates adding a setter reference to fields.
func ExampleAdd_fieldName() {
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    something: 3
spec:
  replicas: 3
`

	object := yaml.MustParse(deployment) // parse the configuration
	err := object.PipeE(&Add{
		Ref:       "#/definitions/io.k8s.cli.setters.replicas",
		FieldName: "replicas",
	})
	if err != nil {
		panic(err)
	}

	// Print the object with the update value
	fmt.Println(object.MustString())

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: nginx-deployment
	//   annotations:
	//     something: 3
	// spec:
	//   replicas: 3 # {"$openapi":"replicas"}
}

// ExampleAdd demonstrates adding a setter reference to fields.
func ExampleAdd_fieldValue() {
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    something: 3
spec:
  replicas: 3
`

	object := yaml.MustParse(deployment) // parse the configuration
	err := object.PipeE(&Add{
		Ref:        "#/definitions/io.k8s.cli.setters.replicas",
		FieldValue: "3",
	})
	if err != nil {
		panic(err)
	}

	// Print the object with the update value
	fmt.Println(object.MustString())

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: nginx-deployment
	//   annotations:
	//     something: 3 # {"$openapi":"replicas"}
	// spec:
	//   replicas: 3 # {"$openapi":"replicas"}
}
