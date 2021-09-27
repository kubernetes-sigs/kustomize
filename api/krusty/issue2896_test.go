// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func writeIssueBase(th kusttest_test.Harness) {
	th.WriteK("base", `
nameSuffix: -test-api

resources:
 - deploy.yaml
`)
	th.WriteF("base/deploy.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    spec:
      containers:
        - name: example
          image: example:1.0
          volumeMounts:
            - name: conf
              mountPath: /etc/config
      volumes:
        - name: conf
          configMap:
            name: conf
`)
}

func TestIssue2896Base(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeIssueBase(th)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-test-api
spec:
  template:
    spec:
      containers:
      - image: example:1.0
        name: example
        volumeMounts:
        - mountPath: /etc/config
          name: conf
      volumes:
      - configMap:
          name: conf
        name: conf
`)
}

func TestIssue2896Overlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeIssueBase(th)
	th.WriteK("overlay", `
resources:
  - ../base

patches:
  - patch: |-
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: example
      spec:
        template:
          spec:
            containers:
             - name: example
               image: example:2.0
`)

	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-test-api
spec:
  template:
    spec:
      containers:
      - image: example:2.0
        name: example
        volumeMounts:
        - mountPath: /etc/config
          name: conf
      volumes:
      - configMap:
          name: conf
        name: conf
`)
}
