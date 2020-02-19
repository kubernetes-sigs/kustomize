// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ExampleSet demonstrates using Set to replace the current field value in an object
func ExampleSet() {
	openapi.ResetOpenAPI()

	// OpenAPI definitions with setter extensions on definitions
	schema := `
{
  "definitions": {
    "io.k8s.cli.setters.replicas": {
      "x-k8s-cli": {
        "setter": {
          "name": "replicas",
          "value": "4"
        }
      }
    }
  }
}
`
	// Resource with field referencing OpenAPI definition
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
`
	_, err := openapi.AddSchema([]byte(schema)) // add the schema definitions
	if err != nil {
		panic(err)
	}
	object := yaml.MustParse(deployment)       // parse the configuration
	err = object.PipeE(&Set{Name: "replicas"}) // set replicas from the setter
	if err != nil {
		panic(err)
	}

	fmt.Println(object.MustString())

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: nginx-deployment
	// spec:
	//   replicas: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
}

// ExampleSet_Substitution demonstrates using Set to substitute a value into the field of
// an object.  Only part of the field value is modified.
func ExampleSet_substitution() {
	openapi.ResetOpenAPI()

	// set the version setter
	schema := `
{
  "definitions": {
    "io.k8s.cli.setters.version": {
      "x-k8s-cli": {
        "setter": {
          "name": "version",
          "value": "1.8.1"
        }
      }
    },
    "io.k8s.cli.substitutions.image": {
      "x-k8s-cli": {
        "substitution": {
          "name": "image",
          "pattern": "nginx:VERSION",
          "values": [
            {"marker": "VERSION", "ref": "#/definitions/io.k8s.cli.setters.version"}
          ]
        }
      }
    }
  }
}`

	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
`

	_, err := openapi.AddSchema([]byte(schema)) // add the schema definitions
	if err != nil {
		panic(err)
	}
	object := yaml.MustParse(deployment)      // parse the configuration
	err = object.PipeE(&Set{Name: "version"}) // set replicas from the setter
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
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: nginx
	//         image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image"}
}

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
	//   replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
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
	//     something: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
	// spec:
	//   replicas: 3 # {"$ref":"#/definitions/io.k8s.cli.setters.replicas"}
}
