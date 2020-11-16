// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy

import (
	"bytes"
	"fmt"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// WNodeFactory makes instances of WNode.
//
// These instances in turn adapt
//   sigs.k8s.io/kustomize/kyaml/yaml.RNode
// to implement ifc.Unstructured.
// This factory is meant to implement ifc.KunstructuredFactory.
//
// This implementation should be thin, as both WNode and WNodeFactory must be
// factored away (deleted) along with ifc.Kunstructured in favor of direct use
// of RNode methods upon completion of
// https://github.com/kubernetes-sigs/kustomize/issues/2506.
//
// See also api/krusty/internal/provider/depprovider.go
type WNodeFactory struct {
}

var _ ifc.KunstructuredFactory = (*WNodeFactory)(nil)

func (k *WNodeFactory) SliceFromBytes(bs []byte) ([]ifc.Kunstructured, error) {
	r := kio.ByteReader{OmitReaderAnnotations: true}
	r.Reader = bytes.NewBuffer(bs)
	yamlRNodes, err := r.Read()
	if err != nil {
		return nil, err
	}
	var result []ifc.Kunstructured
	for i := range yamlRNodes {
		rn := yamlRNodes[i]
		meta, err := rn.GetValidatedMetadata()
		if err != nil {
			return nil, err
		}
		if !shouldDropObject(meta) {
			if foundNil, path := rn.HasNilEntryInList(); foundNil {
				return nil, fmt.Errorf("empty item at %v in object %v", path, rn)
			}
			result = append(result, FromRNode(rn))
		}
	}
	return result, nil
}

// shouldDropObject returns true if the resource should not be accumulated.
func shouldDropObject(m yaml.ResourceMeta) bool {
	_, y := m.ObjectMeta.Annotations[konfig.IgnoredByKustomizeResourceAnnotation]
	return y
}

func (k *WNodeFactory) FromMap(m map[string]interface{}) ifc.Kunstructured {
	rn, err := FromMap(m)
	if err != nil {
		// TODO(#WNodeFactory): handle or bubble error"
		panic(err)
	}
	return rn
}

func (k *WNodeFactory) Hasher() ifc.KunstructuredHasher {
	panic("TODO(#WNodeFactory): implement Hasher")
}

func (k *WNodeFactory) MakeConfigMap(
	kvLdr ifc.KvLoader, args *types.ConfigMapArgs) (ifc.Kunstructured, error) {
	panic("TODO(#WNodeFactory): implement MakeConfigMap")
}

func (k *WNodeFactory) MakeSecret(
	kvLdr ifc.KvLoader, args *types.SecretArgs) (ifc.Kunstructured, error) {
	panic("TODO(#WNodeFactory): implement MakeSecret")
}
