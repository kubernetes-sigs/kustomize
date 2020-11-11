// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy

import (
	"fmt"
	"log"
	"strings"

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
	fields := strings.Split(path, ".")
	rn, err := wn.node.Pipe(yaml.Lookup(fields...))
	if err != nil {
		return nil, err
	}
	if rn == nil {
		return nil, NoFieldError{path}
	}
	yn := rn.YNode()

	// If this is an alias node, resolve it
	if yn.Kind == yaml.AliasNode {
		yn = yn.Alias
	}

	// Return value as map for DocumentNode and MappingNode kinds
	if yn.Kind == yaml.DocumentNode || yn.Kind == yaml.MappingNode {
		var result map[string]interface{}
		if err := yn.Decode(&result); err != nil {
			return nil, err
		}
		return result, err
	}

	// Return value as slice for SequenceNode kind
	if yn.Kind == yaml.SequenceNode {
		var result []interface{}
		for _, node := range yn.Content {
			result = append(result, node.Value)
		}
		return result, nil
	}

	// Return value value directly for all other (ScalarNode) kinds
	return yn.Value, nil
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
func (wn *WNode) GetSlice(path string) ([]interface{}, error) {
	value, err := wn.GetFieldValue(path)
	if err != nil {
		return nil, err
	}
	if sliceValue, ok := value.([]interface{}); ok {
		return sliceValue, nil
	}
	return nil, fmt.Errorf("node %s is not a slice", path)
}

// GetSlice implements ifc.Kunstructured.
func (wn *WNode) GetString(path string) (string, error) {
	value, err := wn.GetFieldValue(path)
	if err != nil {
		return "", err
	}
	if v, ok := value.(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("node %s is not a string: %v", path, value)
}

// Map implements ifc.Kunstructured.
func (wn *WNode) Map() map[string]interface{} {
	var result map[string]interface{}
	if err := wn.node.YNode().Decode(&result); err != nil {
		// Log and die since interface doesn't allow error.
		log.Fatalf("failed to decode ynode: %v", err)
	}
	return result
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
func (wn *WNode) SetAnnotations(annotations map[string]string) {
	wn.setField(yaml.NewMapRNode(&annotations), yaml.MetadataField, yaml.AnnotationsField)
}

// SetGvk implements ifc.Kunstructured.
func (wn *WNode) SetGvk(gvk resid.Gvk) {
	wn.setField(yaml.NewScalarRNode(gvk.Kind), yaml.KindField)
	wn.setField(yaml.NewScalarRNode(fmt.Sprintf("%s/%s", gvk.Group, gvk.Version)), yaml.APIVersionField)
}

// SetLabels implements ifc.Kunstructured.
func (wn *WNode) SetLabels(labels map[string]string) {
	wn.setField(yaml.NewMapRNode(&labels), yaml.MetadataField, yaml.LabelsField)
}

// SetName implements ifc.Kunstructured.
func (wn *WNode) SetName(name string) {
	wn.setField(yaml.NewScalarRNode(name), yaml.MetadataField, yaml.NameField)
}

// SetNamespace implements ifc.Kunstructured.
func (wn *WNode) SetNamespace(ns string) {
	wn.setField(yaml.NewScalarRNode(ns), yaml.MetadataField, yaml.NamespaceField)
}

func (wn *WNode) setField(value *yaml.RNode, path ...string) {
	err := wn.node.PipeE(
		yaml.LookupCreate(yaml.MappingNode, path[0:len(path)-1]...),
		yaml.SetField(path[len(path)-1], value),
	)
	if err != nil {
		// Log and die since interface doesn't allow error.
		log.Fatalf("failed to set field %v: %v", path, err)
	}
}

// UnmarshalJSON implements ifc.Kunstructured.
func (wn *WNode) UnmarshalJSON(data []byte) error {
	return wn.node.UnmarshalJSON(data)
}

type NoFieldError struct {
	Field string
}

func (e NoFieldError) Error() string {
	return fmt.Sprintf("no field named '%s'", e.Field)
}
