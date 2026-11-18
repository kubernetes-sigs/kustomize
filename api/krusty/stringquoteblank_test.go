// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// This test verifies that long lines are NOT wrapped in YAML output.
// See https://github.com/kubernetes-sigs/kustomize/issues/947
func TestLongLineBreaks(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test
spec:
  template:
    spec:
      containers:
      - name: mariadb
        image: test
        env:
        - name: SHORT_STRING
          value: short_string
        - name: SHORT_STRING_WITH_SINGLE_QUOTE
          value: 'short_string'
        - name: SHORT_STRING_WITH_DOUBLE_QUOTE
          value: "short_string"
        - name: SHORT_STRING_BLANK
          value: short string
        - name: LONG_STRING_BLANK
          value: Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius.
        - name: LONG_STRING_BLANK_WITH_SINGLE_QUOTE
          value: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius.'
        - name: LONG_STRING_BLANK_WITH_DOUBLE_QUOTE
          value: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius."
        - name: INVALID_PLAIN_STYLE_STRING
          value: ': test'
`)
	th.WriteK(".", `
resources:
- deployment.yaml
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test
spec:
  template:
    spec:
      containers:
      - name: mariadb
        image: test
        env:
        - name: SHORT_STRING
          value: short_string
        - name: SHORT_STRING_WITH_SINGLE_QUOTE
          value: 'short_string'
        - name: SHORT_STRING_WITH_DOUBLE_QUOTE
          value: "short_string"
        - name: SHORT_STRING_BLANK
          value: short string
        - name: LONG_STRING_BLANK
          value: Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius.
        - name: LONG_STRING_BLANK_WITH_SINGLE_QUOTE
          value: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius.'
        - name: LONG_STRING_BLANK_WITH_DOUBLE_QUOTE
          value: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius."
        - name: INVALID_PLAIN_STYLE_STRING
          value: ': test'
`)
}
