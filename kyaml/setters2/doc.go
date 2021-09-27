// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
// Package setters2 contains libraries for setting resource field values from OpenAPI setter
// extensions.
//
// Setters
//
// Setters are used to programmatically set configuration field values -- e.g. through a cli or ui.
//
// Setters are defined through OpenAPI definitions using the x-k8s-cli extension.
// Note: additional OpenAPI definitions may be registered through openapi.AddSchema([]byte)
//
// Example OpenAPI schema containing a setter:
//
//  {
//    "definitions": {
//      "io.k8s.cli.setters.replicas": {
//        "x-k8s-cli": {
//          "setter": {
//            "name": "replicas",
//            "value": "4"
//          }
//        }
//      }
//    }
//  }
//
// Setter fields:
//
//   x-k8s-cli.setter.name: name of the setter
//   x-k8s-cli.setter.value: value of the setter that should be applied to fields
//
// The setter definition key must be of the form "io.k8s.cli.setters.NAME", where NAME matches the
// value of "x-k8s-cli.setter.name".
//
// When Set.Filter is called, the named setter will have its value applied to all resource
// fields referencing it.
//
// Fields may reference setters through a yaml comment containing the serialized JSON OpenAPI.
//
// Example Deployment resource with a "spec.replicas" field set by the "replicas" setter:
//
//   apiVersion: apps/v1
//   kind: Deployment
//   metadata:
//     name: nginx-deployment
//   spec:
//     replicas: 4 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
//
// If the OpenAPI io.k8s.cli.setters.replicas x-k8s-cli.setter.value was changed from "4" to "5",
// then calling Set{Name: "replicas"}.Filter(deployment) would update the Deployment spec.replicas
// value from 4 to 5.
//
// Updated OpenAPI:
//
//  {
//    "definitions": {
//      "io.k8s.cli.setters.replicas": {
//        "x-k8s-cli": {
//          "setter": {
//            "name": "replicas",
//            "value": "5"
//          }
//        }
//      }
//    }
//  }
//
// Updated Deployment Configuration:
//
//   apiVersion: apps/v1
//   kind: Deployment
//   metadata:
//     name: nginx-deployment
//   spec:
//     replicas: 5 # {"$ref": "#/definitions/io.k8s.cli.setters.replicas"}
//
// Substitutions
//
// Substitutions are used to programmatically set configuration field values using multiple
// setters which are substituted into a pattern string.
//
// Substitutions may be used when a field value does not cleanly map to a single setter, but
// instead matches some string pattern where setters may be substituted in.
//
// Fields may reference substitutions the same way they do setters, however substitutions
// reference setters from which they are derived.
//
// Example OpenAPI schema containing a substitution derived from 2 setters:
//
//   {
//     "definitions": {
//       "io.k8s.cli.setters.image-name": {
//         "x-k8s-cli": {
//           "setter": {
//             "name": "image-name",
//             "value": "nginx"
//           }
//         }
//       },
//       "io.k8s.cli.setters.image-tag": {
//         "x-k8s-cli": {
//           "setter": {
//             "name": "image-tag",
//             "value": "1.8.1"
//           }
//         }
//       },
//       "io.k8s.cli.substitutions.image-name-tag": {
//         "x-k8s-cli": {
//           "substitution": {
//             "name": "image-name-tag",
//             "pattern": "IMAGE_NAME:IMAGE_TAG",
//             "values": [
//                 {"marker": "IMAGE_NAME", "ref": "#/definitions/io.k8s.cli.setters.image-name"}
//                 {"marker": "IMAGE_TAG", "ref": "#/definitions/io.k8s.cli.setters.image-tag"}
//             ]
//           }
//         }
//       }
//     }
//   }
//
// Substitution Fields.
//
//  x-k8s-cli.substitution.name: name of the substitution
//  x-k8s-cli.substitution.pattern: string pattern to substitute markers into
//  x-k8s-cli.substitution.values.marker: the marker substring within pattern to replace
//  x-k8s-cli.substitution.values.ref: the setter ref containing the value to replace the marker with
//
// The substitution is composed of a "pattern" containing markers, and a list of setter "values"
// which are substituted into the markers.
//
// Example Deployment with substitution:
//
//   apiVersion: apps/v1
//   kind: Deployment
//   metadata:
//     name: nginx-deployment
//   spec:
//     template:
//       spec:
//         containers:
//         - name: nginx
//           image: nginx:1.8.1 # {"$ref": "#/definitions/io.k8s.cli.substitutions.image-name-tag"}
//
// spec.template.spec.containers[name=nginx].image is set by the "image" substitution any time
// either "image-name" or "image-tag" is set.  Whenever any setter referenced by a substitution
// is set, the substitution will be recalculated by substituting its values into its pattern.
//
//
// If the OpenAPI io.k8s.cli.setters.image-name x-k8s-cli.setter.value was changed from "1.8.1"
// to "1.8.2", then calling either Set{Name: "image-name"}.Filter(deployment) or
// Set{Name: "image-tag"}.Filter(deployment) would update the Deployment field
// spec.template.spec.container[name=nginx].image from "nginx:1.8.1" to "nginx:1.8.2".
//
// Adding Field References
//
// References to setters and substitutions may be added to fields using the Add Filter.
// Add will write a JSON OpenAPI string as a comment to any fields matching the specified
// FieldName add FieldValue.
package setters2
