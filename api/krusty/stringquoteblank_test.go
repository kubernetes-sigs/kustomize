package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// This test is for output string style.
// Currently all quotes will be removed if the string is valid as plain (unquoted) style.
// If a string cannot be unquoted, it will be put into a pair of single quotes.
// See https://yaml.org/spec/1.2/spec.html#id2788859 for more details about what kind of string
// is invalid as plain style.
func TestLongLineBreaks(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("/app/deployment.yaml", `
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
	th.WriteK("/app", `
resources:
- deployment.yaml
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test
spec:
  template:
    spec:
      containers:
      - env:
        - name: SHORT_STRING
          value: short_string
        - name: SHORT_STRING_WITH_SINGLE_QUOTE
          value: short_string
        - name: SHORT_STRING_WITH_DOUBLE_QUOTE
          value: short_string
        - name: SHORT_STRING_BLANK
          value: short string
        - name: LONG_STRING_BLANK
          value: Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius.
        - name: LONG_STRING_BLANK_WITH_SINGLE_QUOTE
          value: Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius.
        - name: LONG_STRING_BLANK_WITH_DOUBLE_QUOTE
          value: Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas suscipit ex non molestie varius.
        - name: INVALID_PLAIN_STYLE_STRING
          value: ': test'
        image: test
        name: mariadb
`)
}
