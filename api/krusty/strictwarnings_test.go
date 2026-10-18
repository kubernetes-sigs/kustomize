// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func writeDeployWithCommonLabelsWarnings(th kusttest_test.Harness) {
	th.WriteK(".", `
commonLabels:
  foo: bar
resources:
- deploy.yaml
`)
	th.WriteF("deploy.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  numReplicas: 1
`)
}

func writeDeployWithoutWarnings(th kusttest_test.Harness) {
	th.WriteK(".", `
resources:
- deploy.yaml
`)
	th.WriteF("deploy.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  numReplicas: 1
`)
}

func TestEnabledStrictWarningsLabel(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeDeployWithCommonLabelsWarnings(th)

	opts := th.MakeDefaultOptions()
	opts.StrictWarnings = true
	got := th.RunWithErr(".", opts)
	wanted := "# Warning: 'commonLabels'"
	if !strings.Contains(got.Error(), wanted) {
		t.Errorf("got %v not expected: %s", got, wanted)
	}
}

func TestEnabledStrictWarningsLabelWithoutWarnings(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeDeployWithoutWarnings(th)

	opts := th.MakeDefaultOptions()
	opts.StrictWarnings = true
	got := th.Run(".", opts)
	th.AssertActualEqualsExpected(got, `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  numReplicas: 1
`)
}

func TestNotEnabledStrictWarningsLabelAndWarnings(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeDeployWithCommonLabelsWarnings(th)

	opts := th.MakeDefaultOptions()
	opts.StrictWarnings = false
	got := th.Run(".", opts)
	th.AssertActualEqualsExpected(got, `
apiVersion: v1
kind: Deployment
metadata:
  labels:
    foo: bar
  name: storefront
spec:
  numReplicas: 1
  selector:
    matchLabels:
      foo: bar
  template:
    metadata:
      labels:
        foo: bar
`)
}

func TestNotEnabledStrictWarningsLabelAndWithoutWarnings(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeDeployWithoutWarnings(th)

	opts := th.MakeDefaultOptions()
	opts.StrictWarnings = false
	got := th.Run(".", opts)
	th.AssertActualEqualsExpected(got, `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  numReplicas: 1
`)
}
