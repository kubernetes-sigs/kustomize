// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// MakeVaultSecret makes generates a config map ref based on a vault secret path.
//
// ConfigMaps and Secrets are similar to vault secrets.
//
// A VaultSecret has string field `path ` which holds an identifier, used by the
// client, to choose the algorithm to interpret the `data` field.  Kubernetes
func MakeVaultSecret(ldr ifc.KvLoader, args *types.VaultSecretArgs) (rn *yaml.RNode, err error) {
	rn, err = makeBaseNode("ConfigMap", args.Name, args.Namespace)
	if err != nil {
		return nil, err
	}

	m, err := makeValidatedDataMap(ldr, args.Name, args.KvPairSources)
	if err != nil {
		return nil, err
	}
	if err = rn.LoadMapIntoConfigMapData(m); err != nil {
		return nil, err
	}
	if err := copyLabelsAndAnnotations(rn, args.Options); err != nil {
		return nil, err
	}
	if err := setImmutable(rn, args.Options); err != nil {
		return nil, err
	}
	return rn, nil
}
