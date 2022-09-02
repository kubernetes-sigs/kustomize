// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// TestPropFile tests including prop files
func TestPropFile(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	// Creates a prop file with lf linebreaks (default)
	th.WriteF("testPropLf.prop", `# Comment
ssl.truststore=/truststore.jks
ssl.keystore=/keystore.jks
ssl.keystore.password.file=/password.raw
`)

	// Creates a prop file with Windows (cr-lf) linebreaks
	th.WriteF("testPropCrLf.prop", strings.Replace(`# Comment
ssl.truststore=/truststore.jks
ssl.keystore=/keystore.jks
ssl.keystore.password.file=/password.raw
`, "\n", "\r\n", -1))

	// Create a simple kustomize file which uses the above prop files
	th.WriteK(".", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
  - name: prop
    files:
    - testPropLf.prop
    - testPropCrLf.prop
`)
	m := th.Run(".", th.MakeDefaultOptions())

	// The test asserts that both prop files (with lf and crlf)
	// are included the same way
	th.AssertActualEqualsExpected(m, `apiVersion: v1
data:
  testPropCrLf.prop: |
    # Comment
    ssl.truststore=/truststore.jks
    ssl.keystore=/keystore.jks
    ssl.keystore.password.file=/password.raw
  testPropLf.prop: |
    # Comment
    ssl.truststore=/truststore.jks
    ssl.keystore=/keystore.jks
    ssl.keystore.password.file=/password.raw
kind: ConfigMap
metadata:
  name: prop-mb4b77m68g
`)
}
