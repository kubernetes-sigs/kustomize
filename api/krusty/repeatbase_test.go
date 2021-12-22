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
// base
// |    deployment.yaml
// |    kustomization.yaml
// canary
// |    deployment-canary-patch.yaml
// |    kustomization.yaml
// mango
// |    deployment-mango-patch.yaml
// |    deployment-mango-canary-patch.yaml
// |    kustomization.yaml
func TestRepeatBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
  - deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: banana
spec:
  selector:
    matchLabels:
      component: banana
  template:
    metadata:
      labels:
        component: banana
    spec:
      containers:
      - name: banana
        image: image
`)

	th.WriteK("canary", `
resources:
  - ../base
patches:
- patch: |
    - op: replace
      path: /metadata/name
      value: banana-canary
  target: 
    kind: Deployment
- path: deployment-canary-patch.yaml
`)
	th.WriteF("canary/deployment-canary-patch.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: banana-canary
  labels:
    type: canary
spec:
  selector:
    matchLabels:
      component: banana
      type: canary
  template:
    metadata:
      labels:
        component: banana
        type: canary
    spec:
      containers:
      - name: banana
        image: image-canary
`)

	th.WriteK("mango", `
nameSuffix: -mango
resources:
  - ../base
  - ../canary
patches:
- path: deployment-mango-base-patch.yaml
- path: deployment-mango-canary-patch.yaml
`)
	th.WriteF("mango/deployment-mango-base-patch.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: banana
spec:
  template:
    spec:
      containers:
      - name: banana
        image: image-mango
`)
	th.WriteF("mango/deployment-mango-canary-patch.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: banana-canary
spec:
  template:
    spec:
      containers:
      - name: banana
        image: image-canary-mango
`)

	err := th.RunWithErr("mango", th.MakeDefaultOptions())
	if !strings.Contains(
		err.Error(), "multiple matches for Id Deployment.v1.apps/banana.[noNs]; failed to find unique target for patch Deployment.v1.apps/banana.[noNs]") {
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
