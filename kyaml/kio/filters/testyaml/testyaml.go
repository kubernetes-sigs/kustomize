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

var UnformattedYaml3 = []byte(`
apiVersion: v1
items:
- apiVersion: v1
  kind: Namespace
  metadata:
    name: server-dev
    resourceVersion: "7881"
    selfLink: /api/v1/namespaces/server-dev
  status:
    phase: Active
- apiVersion: v1
  kind: Namespace
  metadata:
    name: kube-node-lease
    resourceVersion: "40"
    selfLink: /api/v1/namespaces/kube-node-lease
  status:
    phase: Active
- apiVersion: v1
  kind: Namespace
  metadata:
    name: kube-public
    resourceVersion: "26"
    selfLink: /api/v1/namespaces/kube-public
  status:
    phase: Active
- apiVersion: v1
  kind: Namespace
  metadata:
    name: kube-system
    resourceVersion: "143"
    selfLink: /api/v1/namespaces/kube-system
  status:
    phase: Active
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
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

var FormattedJSON1 = []byte(`{
  "apiVersion": "example.com/v1beta1",
  "kind": "MyType",
  "spec": "a",
  "status": {
    "conditions": [
      3,
      1,
      2
    ]
  }
}
`)

var FormattedYaml3 = []byte(`apiVersion: v1
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
items:
- apiVersion: v1
  kind: Namespace
  metadata:
    name: server-dev
    resourceVersion: "7881"
    selfLink: /api/v1/namespaces/server-dev
  status:
    phase: Active
- apiVersion: v1
  kind: Namespace
  metadata:
    name: kube-node-lease
    resourceVersion: "40"
    selfLink: /api/v1/namespaces/kube-node-lease
  status:
    phase: Active
- apiVersion: v1
  kind: Namespace
  metadata:
    name: kube-public
    resourceVersion: "26"
    selfLink: /api/v1/namespaces/kube-public
  status:
    phase: Active
- apiVersion: v1
  kind: Namespace
  metadata:
    name: kube-system
    resourceVersion: "143"
    selfLink: /api/v1/namespaces/kube-system
  status:
    phase: Active
`)
