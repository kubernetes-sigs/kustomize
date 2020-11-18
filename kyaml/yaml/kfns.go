// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/errors"
)

// AnnotationClearer removes an annotation at metadata.annotations.
// Returns nil if the annotation or field does not exist.
type AnnotationClearer struct {
	Kind string `yaml:"kind,omitempty"`
	Key  string `yaml:"key,omitempty"`
}

func (c AnnotationClearer) Filter(rn *RNode) (*RNode, error) {
	return rn.Pipe(
		PathGetter{Path: []string{MetadataField, AnnotationsField}},
		FieldClearer{Name: c.Key})
}

func ClearAnnotation(key string) AnnotationClearer {
	return AnnotationClearer{Key: key}
}

// ClearEmptyAnnotations clears the keys, annotations
// and metadata if they are empty/null
func ClearEmptyAnnotations(rn *RNode) error {
	_, err := rn.Pipe(Lookup(MetadataField), FieldClearer{
		Name: AnnotationsField, IfEmpty: true})
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = rn.Pipe(FieldClearer{Name: MetadataField, IfEmpty: true})
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

// k8sDataSetter place key value pairs in either a 'data' or 'binaryData' field.
// Useful for creating ConfigMaps and Secrets.
type k8sDataSetter struct {
	Key             string `yaml:"key,omitempty"`
	Value           string `yaml:"value,omitempty"`
	ProtectExisting bool   `yaml:"protectExisting,omitempty"`
}

func (s k8sDataSetter) Filter(rn *RNode) (*RNode, error) {
	if !utf8.Valid([]byte(s.Value)) {
		// Core k8s ConfigMaps store k,v pairs with 'v' passing the above utf8
		// test in a mapping field called "data" as a string. Pairs with a 'v'
		// failing this test go into a field called binaryData as a []byte.
		// TODO: support this distinction in kyaml with NodeTagBytes?
		return nil, fmt.Errorf(
			"key '%s' appears to have non-utf8 data; "+
				"binaryData field not yet supported", s.Key)
	}
	keyNode, err := rn.Pipe(Lookup(DataField, s.Key))
	if err != nil {
		return nil, err
	}
	if keyNode != nil && s.ProtectExisting {
		return nil, fmt.Errorf(
			"protecting existing %s='%s' against attempt to add new value '%s'",
			s.Key, strings.TrimSpace(keyNode.MustString()), s.Value)
	}
	v := NewScalarRNode(s.Value)
	v.YNode().Tag = NodeTagString
	// Add quotes?
	// v.YNode().Style = yaml.SingleQuotedStyle
	_, err = rn.Pipe(
		LookupCreate(yaml.MappingNode, DataField), SetField(s.Key, v))
	return rn, err
}

func SetK8sData(key, value string) k8sDataSetter {
	return k8sDataSetter{Key: key, Value: value, ProtectExisting: true}
}

// k8sMetaSetter sets a name at metadata.{key}.
// Creates metadata if does not exist.
type k8sMetaSetter struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (s k8sMetaSetter) Filter(rn *RNode) (*RNode, error) {
	v := NewScalarRNode(s.Value)
	v.YNode().Tag = NodeTagString
	_, err := rn.Pipe(
		PathGetter{Path: []string{MetadataField}, Create: yaml.MappingNode},
		FieldSetter{Name: s.Key, Value: v})
	return rn, err
}

func SetK8sName(value string) k8sMetaSetter {
	return k8sMetaSetter{Key: NameField, Value: value}
}

func SetK8sNamespace(value string) k8sMetaSetter {
	return k8sMetaSetter{Key: NamespaceField, Value: value}
}

// AnnotationSetter sets an annotation at metadata.annotations.
// Creates metadata.annotations if does not exist.
type AnnotationSetter struct {
	Kind  string `yaml:"kind,omitempty"`
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (s AnnotationSetter) Filter(rn *RNode) (*RNode, error) {
	// some tools get confused about the type if annotations are not quoted
	v := NewScalarRNode(s.Value)
	v.YNode().Tag = NodeTagString
	v.YNode().Style = yaml.SingleQuotedStyle

	if err := ClearEmptyAnnotations(rn); err != nil {
		return nil, err
	}

	return rn.Pipe(
		PathGetter{
			Path:   []string{MetadataField, AnnotationsField},
			Create: yaml.MappingNode},
		FieldSetter{Name: s.Key, Value: v})
}

func SetAnnotation(key, value string) AnnotationSetter {
	return AnnotationSetter{Key: key, Value: value}
}

// AnnotationGetter gets an annotation at metadata.annotations.
// Returns nil if metadata.annotations does not exist.
type AnnotationGetter struct {
	Kind  string `yaml:"kind,omitempty"`
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// AnnotationGetter returns the annotation value.
// Returns "", nil if the annotation does not exist.
func (g AnnotationGetter) Filter(rn *RNode) (*RNode, error) {
	v, err := rn.Pipe(
		PathGetter{Path: []string{MetadataField, AnnotationsField, g.Key}})
	if v == nil || err != nil {
		return v, err
	}
	if g.Value == "" || v.value.Value == g.Value {
		return v, err
	}
	return nil, err
}

func GetAnnotation(key string) AnnotationGetter {
	return AnnotationGetter{Key: key}
}

// LabelSetter sets a label at metadata.labels.
// Creates metadata.labels if does not exist.
type LabelSetter struct {
	Kind  string `yaml:"kind,omitempty"`
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (s LabelSetter) Filter(rn *RNode) (*RNode, error) {
	// some tools get confused about the type if labels are not quoted
	v := NewScalarRNode(s.Value)
	v.YNode().Tag = NodeTagString
	v.YNode().Style = yaml.SingleQuotedStyle
	return rn.Pipe(
		PathGetter{
			Path: []string{MetadataField, LabelsField}, Create: yaml.MappingNode},
		FieldSetter{Name: s.Key, Value: v})
}

func SetLabel(key, value string) LabelSetter {
	return LabelSetter{Key: key, Value: value}
}
