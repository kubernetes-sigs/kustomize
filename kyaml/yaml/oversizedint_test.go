// Copyright the Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestIsOversizedYAMLInteger(t *testing.T) {
	cases := []struct {
		value    string
		expected bool
	}{
		// Fits in uint64, including 2^64-1 with and without padding.
		{value: "0", expected: false},
		{value: "255", expected: false},
		{value: "0xFF", expected: false},
		{value: "0xFFFFFFFFFFFFFFFF", expected: false},
		{value: "0x000000000000000000000000FFFFFFFFFFFFFFFF", expected: false},
		{value: "18446744073709551615", expected: false},
		{value: "-9223372036854775808", expected: false},
		// 2^64 does not fit. Decimal is still an integer literal; hex must be too.
		{value: "18446744073709551616", expected: true},
		{value: "0x10000000000000000", expected: true},
		{value: "0x0000000000000000000000010000000000000000", expected: true},
		{value: "0X10000000000000000", expected: true},
		{value: "+0x10000000000000000", expected: true},
		{value: "-0x10000000000000000", expected: true},
		{value: "0x1_0000_0000_0000_0000", expected: true},
		{value: "0b10000000000000000000000000000000000000000000000000000000000000000", expected: true},
		{value: "0o2000000000000000000000", expected: true},
		// Not integer literals.
		{value: "hello", expected: false},
		{value: "0xgg", expected: false},
		{value: "0x", expected: false},
		{value: `"0x10000000000000000"`, expected: false},
		{value: "true", expected: false},
		{value: "1.5", expected: false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.expected, yaml.IsOversizedYAMLInteger(tc.value), tc.value)
	}
}

func TestIsYaml1_1NonStringOversizedHex(t *testing.T) {
	node := func(v string) *yaml.Node {
		return &yaml.Node{Kind: yaml.ScalarNode, Value: v}
	}
	// 2^64-1 still resolves as an integer, so it was already non-string.
	assert.True(t, yaml.IsYaml1_1NonString(node("0xFFFFFFFFFFFFFFFF")))
	// 2^64 hex does not resolve in go-yaml and used to be treated as a plain string.
	assert.True(t, yaml.IsYaml1_1NonString(node("0x10000000000000000")))
	assert.False(t, yaml.IsYaml1_1NonString(node("wallet")))
}

func TestQuoteOversizedIntegerScalars(t *testing.T) {
	in := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  annotations:
    addr: 0x10000000000000000
data:
  A: "0x000000000000000000000000FFFFFFFFFFFFFFFF"
  B: 0x0000000000000000000000010000000000000000
  D: 0x10000000000000000
  plain: hello
  fit: "255"
binaryData:
  blob: 0x10000000000000000
`
	got, err := yaml.QuoteOversizedIntegerScalars([]byte(in))
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  annotations:
    addr: 0x10000000000000000
data:
  A: "0x000000000000000000000000FFFFFFFFFFFFFFFF"
  B: "0x0000000000000000000000010000000000000000"
  D: "0x10000000000000000"
  plain: hello
  fit: "255"
binaryData:
  blob: "0x10000000000000000"
`, string(got))
}

func TestQuoteOversizedIntegerScalarsSecretAndIgnoredKinds(t *testing.T) {
	secret := []byte(`apiVersion: v1
kind: Secret
metadata:
  name: s
stringData:
  D: 0x10000000000000000
data:
  C: "0xFF"
`)
	got, err := yaml.QuoteOversizedIntegerScalars(secret)
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  name: s
stringData:
  D: "0x10000000000000000"
data:
  C: "0xFF"
`, string(got))

	other := []byte(`apiVersion: v1
kind: Config
metadata:
  name: not-a-configmap
data:
  D: 0x10000000000000000
`)
	got, err = yaml.QuoteOversizedIntegerScalars(other)
	require.NoError(t, err)
	assert.Equal(t, string(other), string(got))
}

func TestQuoteOversizedIntegerScalarsFlowAndComment(t *testing.T) {
	in := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test-config\ndata: {B: 0x10000000000000000, D: 0x10000000000000000} # keep\n")
	got, err := yaml.QuoteOversizedIntegerScalars(in)
	require.NoError(t, err)
	assert.Equal(t, "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test-config\ndata: {B: \"0x10000000000000000\", D: \"0x10000000000000000\"} # keep\n", string(got))
}
