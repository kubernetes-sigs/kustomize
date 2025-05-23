// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// SecretArgs contains the metadata of how to generate a secret.
type SecretArgs struct {
	// GeneratorArgs for the secret.
	GeneratorArgs `json:",inline,omitempty" yaml:",inline,omitempty"`

	// Type of the secret.
	//
	// This is the same field as the secret type field in v1/Secret:
	// It can be "Opaque" (default), or "kubernetes.io/tls".
	//
	// If type is "kubernetes.io/tls", then "literals" or "files" must have exactly two
	// keys: "tls.key" and "tls.crt"
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// StringData if true generates a v1/Secret with plain-text stringData fields
	// instead of base64-encoded data fields. If any fields are not UTF-8, they
	// are still base64-encoded and stored as data as a fallback behavior. This
	// is similar to the default behavior of a ConfigMap.
	StringData bool `json:"stringData,omitempty" yaml:"stringData,omitempty"`
}
