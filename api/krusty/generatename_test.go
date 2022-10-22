// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Coverage for issue #641
func TestGenerateName(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
resources:
- job.yaml
namePrefix: pre-
nameSuffix: -post
`)
	th.WriteF("job.yaml", `
apiVersion: batch/v1
kind: Job
metadata:
  generateName: job-
spec:
  template:
    spec:
      containers:
      - name: job
        image: run/job:1.0
        command:
        - echo
        - done
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: batch/v1
kind: Job
metadata:
  generateName: job-
spec:
  template:
    spec:
      containers:
      - command:
        - echo
        - done
        image: run/job:1.0
        name: job
`)
}
