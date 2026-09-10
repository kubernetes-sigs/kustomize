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

	// EmitStringData if true generates a v1/Secret with plain-text stringData fields
	// instead of base64-encoded data fields. If a generating field does not have a
	// UTF-8 value, it falls back to being stored as a base64-encoded data field. This
	// is similar to the default binaryData fallback of a ConfigMapGenerator.
	EmitStringData bool `json:"emitStringData,omitempty" yaml:"emitStringData,omitempty"`
}
