// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"
)

func TestIssue596AllowDirectoriesThatAreSubstringsOfEachOther(t *testing.T) {
	th := makeTestHarness(t)
	th.WriteK("/app/base", "")
	th.WriteK("/app/overlays/aws", `
resources:
- ../../base
`)
	th.WriteK("/app/overlays/aws-nonprod", `
resources:
- ../aws
`)
	th.WriteK("/app/overlays/aws-sandbox2.us-east-1", `
resources:
- ../aws-nonprod
`)
	m := th.Run("/app/overlays/aws-sandbox2.us-east-1", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, "")
}
