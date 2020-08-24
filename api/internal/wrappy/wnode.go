// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy

import (
	"log"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// WNode implements ifc.Kunstructured using yaml.RNode.
//
// It exists only to help manage a switch from
// kunstruct.UnstructAdapter to yaml.RNode as the core
// representation of KRM objects in kustomize.
//
// It's got a silly name because we don't want it around for long,
// and want its use to be obvious.
type WNode struct {
	node *yaml.RNode
}

var _ ifc.Kunstructured = (*WNode)(nil)

func NewWNode() *WNode {
	return FromRNode(yaml.NewRNode(nil))
}

func FromRNode(node *yaml.RNode) *WNode {
	return &WNode{node: node}
}

func (wn *WNode) demandMetaData(label string) yaml.ResourceMeta {
	meta, err := wn.node.GetMeta()
	if err != nil {
		// Log and die since interface doesn't allow error.
		log.Fatalf("for %s', expected valid resource: %v", label, err)
	}
	return meta
}

// Copy implements ifc.Kunstructured.
func (wn *WNode) Copy() ifc.Kunstructured {
	return &WNode{node: wn.node.Copy()}
}

// GetAnnotations implements ifc.Kunstructured.
func (wn *WNode) GetAnnotations() map[string]string {
	return wn.demandMetaData("GetAnnotations").Annotations
}

// GetFieldValue implements ifc.Kunstructured.
func (wn *WNode) GetFieldValue(path string) (interface{}, error) {
	// The argument is a json path, e.g. "metadata.name"
	// fields := strings.Split(path, ".")
	// return wn.node.Pipe(yaml.Lookup(fields...))
	panic("TODO(#WNode): GetFieldValue; implement or drop from API")
}

// GetGvk implements ifc.Kunstructured.
func (wn *WNode) GetGvk() resid.Gvk {
	meta := wn.demandMetaData("GetGvk")
	g, v := resid.ParseGroupVersion(meta.APIVersion)
	return resid.Gvk{Group: g, Version: v, Kind: meta.Kind}
}

// GetKind implements ifc.Kunstructured.
func (wn *WNode) GetKind() string {
	return wn.demandMetaData("GetKind").Kind
}

// GetLabels implements ifc.Kunstructured.
func (wn *WNode) GetLabels() map[string]string {
	return wn.demandMetaData("GetLabels").Labels
}

// GetName implements ifc.Kunstructured.
func (wn *WNode) GetName() string {
	return wn.demandMetaData("GetName").Name
}

// GetSlice implements ifc.Kunstructured.
func (wn *WNode) GetSlice(string) ([]interface{}, error) {
	panic("TODO(#WNode) GetSlice; implement or drop from API")
}

// GetSlice implements ifc.Kunstructured.
func (wn *WNode) GetString(string) (string, error) {
	panic("TODO(#WNode) GetString; implement or drop from API")
}

// Map implements ifc.Kunstructured.
func (wn *WNode) Map() map[string]interface{} {
	panic("TODO(#WNode) Map; implement or drop from API")
}

// MarshalJSON implements ifc.Kunstructured.
func (wn *WNode) MarshalJSON() ([]byte, error) {
	return wn.node.MarshalJSON()
}

// MatchesAnnotationSelector implements ifc.Kunstructured.
func (wn *WNode) MatchesAnnotationSelector(string) (bool, error) {
	panic("TODO(#WNode) MatchesAnnotationSelector; implement or drop from API")
}

// MatchesLabelSelector implements ifc.Kunstructured.
func (wn *WNode) MatchesLabelSelector(string) (bool, error) {
	panic("TODO(#WNode) MatchesLabelSelector; implement or drop from API")
}

// SetAnnotations implements ifc.Kunstructured.
func (wn *WNode) SetAnnotations(map[string]string) {
	panic("TODO(#WNode) SetAnnotations; implement or drop from API")
}

// SetGvk implements ifc.Kunstructured.
func (wn *WNode) SetGvk(resid.Gvk) {
	panic("TODO(#WNode) SetGvk; implement or drop from API")
}

// SetLabels implements ifc.Kunstructured.
func (wn *WNode) SetLabels(map[string]string) {
	panic("TODO(#WNode) SetLabels; implement or drop from API")
}

// SetName implements ifc.Kunstructured.
func (wn *WNode) SetName(string) {
	panic("TODO(#WNode) SetName; implement or drop from API")
}

// SetNamespace implements ifc.Kunstructured.
func (wn *WNode) SetNamespace(string) {
	panic("TODO(#WNode) SetNamespace; implement or drop from API")
}

// UnmarshalJSON implements ifc.Kunstructured.
func (wn *WNode) UnmarshalJSON(data []byte) error {
	return wn.node.UnmarshalJSON(data)
}
