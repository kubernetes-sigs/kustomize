// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// I expected configMapRef.name and configMapKeyRef.name to be "mysql-g9ttm44c48" not "mysql"

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestNamePrefixSuffixPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PrefixSuffixTransformer").
		PrepBuiltin("AnnotationsTransformer").
		PrepBuiltin("LabelTransformer")
	defer th.Reset()

	th.WriteF("/app/handlers/kustomization.yaml", `
nameSuffix: -suffix

resources:
  - ./deployment.yaml
`)

	th.WriteF("/app/handlers/deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: short
spec:
  template:
    spec:
      containers:
      - name: handler
`)

	th.WriteF("/app/mysql/kustomization.yaml", `
configMapGenerator:
  - name: mysql
    literals:
      - MYSQL_USER=mysql
      - MYSQL_DATABASE=default
`)

	th.WriteK("/app", `
resources:
  - ./mysql
  - ./handlers

configMapGenerator:
  - name: mysql
    behavior: merge
    literals:
      - MYSQL_DATABASE=db
      - MYSQL_USER=user

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

	m := th.Run("/app", th.MakeDefaultOptions())
	// I expected configMapRef.name and configMapKeyRef.name to be "mysql-g9ttm44c48" not "mysql"
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  MYSQL_DATABASE: db
  MYSQL_USER: user
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: mysql-g9ttm44c48
---
apiVersion: extensions/v1beta1
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
              name: mysql
        envFrom:
        - configMapRef:
            name: mysql
        name: handler
`)
}
