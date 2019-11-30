// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"
)

func TestNullValues(t *testing.T) {
	th := makeTestHarness(t)
	th.WriteF("/app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: example
  name: example
spec:
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - args: null
        image: image
        name: example
`)
	th.WriteF("/app/kustomization.yaml", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
`)
	m := th.Run("/app", th.MakeDefaultOptions())

	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: example
  name: example
spec:
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - args: null
        image: image
        name: example
`)
}
