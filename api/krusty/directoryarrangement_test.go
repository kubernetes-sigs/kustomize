// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const expectedResources = `apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`

func TestIssue596AllowDirectoriesThatAreSubstringsOfEachOther(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("base/service.yaml", expectedResources)
	th.WriteK("base", `resources:
- service.yaml`)
	th.WriteK("overlays/aws", `
resources:
- ../../base
`)
	th.WriteK("overlays/aws-nonprod", `
resources:
- ../aws
`)
	th.WriteK("overlays/aws-sandbox2.us-east-1", `
resources:
- ../aws-nonprod
`)
	m := th.Run("overlays/aws-sandbox2.us-east-1", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expectedResources)
}
