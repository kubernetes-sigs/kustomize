// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// MakeSecret makes a kubernetes Secret.
//
// Secret: https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/secret-v1/
//
// ConfigMaps and Secrets are similar.
//
// Like a ConfigMap, a Secret has a `data` field, but unlike a ConfigMap it has
// no `binaryData` field. Secret also provides a `stringData` field.
//
// A Secret's `data` is assumed to be opaque in nature, and assumed to be
// base64 encoded from its original representation, regardless of whether the
// original data was UTF-8 text or binary.
//
// This encoding provides no secrecy. It's just a neutral, common means to
// represent opaque text and binary data.  Beneath the base64 encoding
// is presumably further encoding under control of the Secret's consumer.
//
// A Secret's `stringData` field is similar to ConfigMap's `data` field.
// `stringData` allows specifying non-binary, UTF-8 secret data in string form.
// It is provided as a write-only input field for convenience.
// All keys and values are merged into the data field on write, overwriting any
// existing values. The stringData field is never output when reading from the API.
//
// A Secret has string field `type` which holds an identifier, used by the
// client, to choose the algorithm to interpret the `data` field.  Kubernetes
// cannot make use of this data; it's up to a controller or some pod's service
// to interpret the value, using `type` as a clue as to how to do this.
func MakeSecret(
	ldr ifc.KvLoader, args *types.SecretArgs) (rn *yaml.RNode, err error) {
	rn, err = makeBaseNode("Secret", args.Name, args.Namespace)
	if err != nil {
		return nil, err
	}
	t := "Opaque"
	if args.Type != "" {
		t = args.Type
	}
	if _, err := rn.Pipe(
		yaml.FieldSetter{
			Name:  "type",
			Value: yaml.NewStringRNode(t)}); err != nil {
		return nil, err
	}
	m, err := makeValidatedDataMap(ldr, args.Name, args.KvPairSources)
	if err != nil {
		return nil, err
	}
	if args.EmitStringData {
		if err = rn.LoadMapIntoSecretStringData(m); err != nil {
			return nil, fmt.Errorf("Failed to load map into Secret stringData: %w", err)
		}
	} else {
		if err = rn.LoadMapIntoSecretData(m); err != nil {
			return nil, fmt.Errorf("Failed to load map into Secret data: %w", err)
		}
	}
	copyLabelsAndAnnotations(rn, args.Options)
	setImmutable(rn, args.Options)
	return rn, nil
}
