// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy

import (
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/types"
)

// WNodeFactory makes instances of WNode.
// These instances in turn adapt
//   sigs.k8s.io/kustomize/kyaml/yaml.RNode
// to implement ifc.Unstructured.
// This factory is meant to implement ifc.KunstructuredFactory.
type WNodeFactory struct {
}

var _ ifc.KunstructuredFactory = (*WNodeFactory)(nil)

func (k *WNodeFactory) SliceFromBytes(bs []byte) ([]ifc.Kunstructured, error) {
	panic("TODO(#WNodeFactory): implement SliceFromBytes")
}

func (k *WNodeFactory) FromMap(m map[string]interface{}) ifc.Kunstructured {
	panic("TODO(#WNodeFactory): implement FromMap")
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
