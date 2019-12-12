// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package testyaml contains test data and libraries for formatting
// Kubernetes configuration
package testyaml

var UnformattedYaml1 = []byte(`
spec: a
status:
  conditions:
  - 3
  - 1
  - 2
apiVersion: example.com/v1beta1
kind: MyType
`)

var UnformattedYaml2 = []byte(`
spec2: a
status2:
  conditions:
  - 3
  - 1
  - 2
apiVersion: example.com/v1beta1
kind: MyType2
`)

var UnformattedJSON1 = []byte(`
{
  "spec": "a",
  "status": {"conditions": [3, 1, 2]},
  "apiVersion": "example.com/v1beta1",
  "kind": "MyType"
}
`)

var FormattedYaml1 = []byte(`apiVersion: example.com/v1beta1
kind: MyType
spec: a
status:
  conditions:
  - 3
  - 1
  - 2
`)

var FormattedYaml2 = []byte(`apiVersion: example.com/v1beta1
kind: MyType2
spec2: a
status2:
  conditions:
  - 3
  - 1
  - 2
`)
