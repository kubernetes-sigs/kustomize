// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Coverage for issue #2609
func TestNamePrefixSuffixPatch(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteF("handlers/kustomization.yaml", `
nameSuffix: -suffix

resources:
- deployment.yaml
`)

	th.WriteF("handlers/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: short
spec:
  template:
    spec:
      containers:
      - name: handler
`)

	th.WriteF("mysql/kustomization.yaml", `
configMapGenerator:
- name: mysql
  literals:
  - MYSQL_USER=default
  - MYSQL_DATABASE=default
  - PORT=3306
`)

	th.WriteK(".", `
resources:
- mysql
- handlers

configMapGenerator:
- name: mysql
  behavior: merge
  literals:
  - MYSQL_DATABASE=db
  - MYSQL_USER=my-user
  - MYSQL_PASSWORD='correct horse battery staple'

patches:
- target:
    kind: Deployment
    name: s.*
  patch: |-
    kind: Deployment
    metadata:
      name: ignored
    spec:
      template:
        spec:
          containers:
          - name: handler
            envFrom:
            - configMapRef:
                name: mysql
            env:
            - valueFrom:
                configMapKeyRef:
                  name: mysql
                  key: MYSQL_DATABASE
`)

	m := th.Run(".", th.MakeDefaultOptions())
	// Per #2609, the desired behavior is for configMapRef.name and configMapKeyRef.name to be "mysql-9792mdchtg" not "mysql"
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  MYSQL_DATABASE: db
  MYSQL_PASSWORD: correct horse battery staple
  MYSQL_USER: my-user
  PORT: "3306"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: mysql-9792mdchtg
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: short-suffix
spec:
  template:
    spec:
      containers:
      - env:
        - valueFrom:
            configMapKeyRef:
              key: MYSQL_DATABASE
              name: mysql-9792mdchtg
        envFrom:
        - configMapRef:
            name: mysql-9792mdchtg
        name: handler
`)
}
