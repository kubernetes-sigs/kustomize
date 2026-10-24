// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// MakeSecret makes a kubernetes Secret.
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
	pairs, err := makeValidatedDataMap(ldr, args.Name, args.KvPairSources)
	if err != nil {
		return nil, err
	}
	yamlPairs := make([]yaml.Pair, len(pairs))
	for i, p := range pairs {
		yamlPairs[i] = yaml.Pair{Key: p.Key, Value: p.Value}
	}
	if err = rn.LoadPairsIntoSecretData(yamlPairs); err != nil {
		return nil, err
	}
	copyLabelsAndAnnotations(rn, args.Options)
	setImmutable(rn, args.Options)
	return rn, nil
}