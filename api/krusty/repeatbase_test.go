// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// This is broken since kustomize v3.9.3.
// See https://github.com/kubernetes-sigs/kustomize/issues/3609 for details.

// Here is a structure of a kustomization of one resource inheriting from
// two bases. One of those bases is shared between the canary base and the
// final resource. This is named canary as it is a simple pattern to
// duplicate a resource that can be used with canary deployments.
//
// deployment-base
// |    deployment.yaml
// |    kustomization.yaml
// deployment-canary
// |    deployment-canary-patch.yaml
// |    kustomization.yaml
// deployment-a
// |    deployment-a-patch.yaml
// |    kustomization.yaml
func TestRepeatBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/deployment-base", `
resources:
  - deployment.yaml
`)
	th.WriteF("/app/deployment-base/deployment.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  selector:
    matchLabels:
      component: deployment
  template:
    metadata:
      labels:
        component: deployment
    spec:
      containers:
      - name: container-a
        image: image
`)

	th.WriteK("/app/deployment-canary", `
resources:
  - ../deployment-base
patches:
- patch: |
    - op: replace
      path: /metadata/name
      value: deployment-canary
  target: 
    kind: Deployment
- path: deployment-canary-patch.yaml
`)
	th.WriteF("/app/deployment-canary/deployment-canary-patch.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-canary
  labels:
    type: canary
spec:
  selector:
    matchLabels:
      component: deployment
      type: canary
  template:
    metadata:
      labels:
        component: deployment
    spec:
      containers:
      - name: container-a
        image: image-canary
`)

	th.WriteK("/app/deployment-a", `
nameSuffix: -a
resources:
  - ../deployment-base
  - ../deployment-canary
patches:
- path: deployment-a-base-patch.yaml
- path: deployment-a-canary-patch.yaml
`)
	th.WriteF("/app/deployment-a/deployment-a-base-patch.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - name: container-a
        image: image-a
`)
	th.WriteF("/app/deployment-a/deployment-a-canary-patch.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-canary
spec:
  template:
    spec:
      containers:
      - name: container-a
        image: image-canary-a
`)

	th.WriteK("/app", `
resources:
  - deployment-a
`)

	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	if !strings.Contains(
		err.Error(), "multiple matches for Id apps_v1_Deployment|~X|deployment; failed to find unique target for patch") {
		t.Fatalf("Unexpected err: %v", err)
	}

	// Uncomenting the following makes it work with kustomize v3.9.2 and bellow.

	//  m := th.Run("/app", th.MakeDefaultOptions())
	//	th.AssertActualEqualsExpected(m, `
	//apiVersion: apps/v1
	//kind: Deployment
	//metadata:
	//  name: deployment-a
	//spec:
	//  selector:
	//    matchLabels:
	//      component: deployment
	//  template:
	//    metadata:
	//      labels:
	//        component: deployment
	//    spec:
	//      containers:
	//      - image: image-a
	//        name: container-a
	//---
	//apiVersion: apps/v1
	//kind: Deployment
	//metadata:
	//  labels:
	//    type: canary
	//  name: deployment-canary-a
	//spec:
	//  selector:
	//    matchLabels:
	//      component: deployment
	//      type: canary
	//  template:
	//    metadata:
	//      labels:
	//        component: deployment
	//    spec:
	//      containers:
	//      - image: image-canary-a
	//        name: container-a
	//`)

}
