// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/sets"
)

const (
	// NullNodeTag is the tag set for a yaml.Document that contains no data -- e.g. it isn't a
	// Map, Slice, Document, etc
	NullNodeTag = "!!null"
)

// NullNode returns a RNode point represents a null; value
func NullNode() *RNode {
	return NewRNode(&Node{Tag: NullNodeTag})
}

// IsMissingOrNull returns true if the RNode is nil or contains and explicitly null value.
func IsMissingOrNull(node *RNode) bool {
	if node == nil || node.YNode() == nil || node.YNode().Tag == NullNodeTag {
		return true
	}
	return false
}

// IsEmpty returns true if the RNode is MissingOrNull, or is either a MappingNode with
// no fields, or a SequenceNode with no elements.
func IsEmpty(node *RNode) bool {
	if node == nil || node.YNode() == nil || node.YNode().Tag == NullNodeTag {
		return true
	}

	if node.YNode().Kind == yaml.MappingNode && len(node.YNode().Content) == 0 {
		return true
	}
	if node.YNode().Kind == yaml.SequenceNode && len(node.YNode().Content) == 0 {
		return true
	}

	return false
}

func IsNull(node *RNode) bool {
	return node != nil && node.YNode() != nil && node.YNode().Tag == NullNodeTag
}

func IsFieldEmpty(node *MapNode) bool {
	if node == nil || node.Value == nil || node.Value.YNode() == nil ||
		node.Value.YNode().Tag == NullNodeTag {
		return true
	}

	if node.Value.YNode().Kind == yaml.MappingNode && len(node.Value.YNode().Content) == 0 {
		return true
	}
	if node.Value.YNode().Kind == yaml.SequenceNode && len(node.Value.YNode().Content) == 0 {
		return true
	}

	return false
}

func IsFieldNull(node *MapNode) bool {
	return node != nil && node.Value != nil && node.Value.YNode() != nil &&
		node.Value.YNode().Tag == NullNodeTag
}

// Parser parses values into configuration.
type Parser struct {
	Kind  string `yaml:"kind,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (p Parser) Filter(_ *RNode) (*RNode, error) {
	d := yaml.NewDecoder(bytes.NewBuffer([]byte(p.Value)))
	o := &RNode{value: &yaml.Node{}}
	return o, d.Decode(o.value)
}

// Parse parses a yaml string into an *RNode
func Parse(value string) (*RNode, error) {
	return Parser{Value: value}.Filter(nil)
}

// TODO(pwittrock): test this
func GetStyle(styles ...string) Style {
	var style Style
	for _, s := range styles {
		switch s {
		case "TaggedStyle":
			style |= TaggedStyle
		case "DoubleQuotedStyle":
			style |= DoubleQuotedStyle
		case "SingleQuotedStyle":
			style |= SingleQuotedStyle
		case "LiteralStyle":
			style |= LiteralStyle
		case "FoldedStyle":
			style |= FoldedStyle
		case "FlowStyle":
			style |= FlowStyle
		}
	}
	return style
}

// MustParse parses a yaml string into an *RNode and panics if there is an error
func MustParse(value string) *RNode {
	v, err := Parser{Value: value}.Filter(nil)
	if err != nil {
		panic(err)
	}
	return v
}

// NewScalarRNode returns a new Scalar *RNode containing the provided scalar value.
func NewScalarRNode(value string) *RNode {
	return &RNode{
		value: &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: value,
		}}
}

// NewListRNode returns a new List *RNode containing the provided scalar values.
func NewListRNode(values ...string) *RNode {
	seq := &RNode{value: &yaml.Node{Kind: yaml.SequenceNode}}
	for _, v := range values {
		seq.value.Content = append(seq.value.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: v,
		})
	}
	return seq
}

// NewRNode returns a new RNode pointer containing the provided Node.
func NewRNode(value *yaml.Node) *RNode {
	return &RNode{value: value}
}

// Filter defines a function to manipulate an individual RNode such as by changing
// its values, or returning a field.
//
// When possible, Filters should be serializable to yaml so that they can be described
// declaratively as data.
//
// Analogous to http://www.linfo.org/filters.html
type Filter interface {
	Filter(object *RNode) (*RNode, error)
}

type FilterFunc func(object *RNode) (*RNode, error)

func (f FilterFunc) Filter(object *RNode) (*RNode, error) {
	return f(object)
}

// RNode provides functions for manipulating Kubernetes Resources
// Objects unmarshalled into *yaml.Nodes
type RNode struct {
	// fieldPath contains the path from the root of the KubernetesObject to
	// this field.
	// Only field names are captured in the path.
	// e.g. a image field in a Deployment would be
	// 'spec.template.spec.containers.image'
	fieldPath []string

	// FieldValue contains the value.
	// FieldValue is always set:
	// field: field value
	// list entry: list entry value
	// object root: object root
	value *yaml.Node

	Match []string
}

// MapNode wraps a field key and value.
type MapNode struct {
	Key   *RNode
	Value *RNode
}

type MapNodeSlice []*MapNode

func (m MapNodeSlice) Keys() []*RNode {
	var keys []*RNode
	for i := range m {
		if m[i] != nil {
			keys = append(keys, m[i].Key)
		}
	}
	return keys
}

func (m MapNodeSlice) Values() []*RNode {
	var values []*RNode
	for i := range m {
		if m[i] != nil {
			values = append(values, m[i].Value)
		} else {
			values = append(values, nil)
		}
	}
	return values
}

// ResourceMeta contains the metadata for a both Resource Type and Resource.
type ResourceMeta struct {
	// APIVersion is the apiVersion field of a Resource
	APIVersion string `yaml:"apiVersion,omitempty"`
	// Kind is the kind field of a Resource
	Kind string `yaml:"kind,omitempty"`
	// ObjectMeta is the metadata field of a Resource
	ObjectMeta `yaml:"metadata,omitempty"`
}

// ObjectMeta contains metadata about a Resource
type ObjectMeta struct {
	// Name is the metadata.name field of a Resource
	Name string `yaml:"name,omitempty"`
	// Namespace is the metadata.namespace field of a Resource
	Namespace string `yaml:"namespace,omitempty"`
	// Labels is the metadata.labels field of a Resource
	Labels map[string]string `yaml:"labels,omitempty"`
	// Annotations is the metadata.annotations field of a Resource.
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// GetIdentifier returns a ResourceIdentifier that includes
// the information needed to uniquely identify a resource in a cluster.
func (m *ResourceMeta) GetIdentifier() ResourceIdentifier {
	return ResourceIdentifier{
		Name:       m.Name,
		Namespace:  m.Namespace,
		APIVersion: m.APIVersion,
		Kind:       m.Kind,
	}
}

// ResourceIdentifier contains the information needed to uniquely
// identify a resource in a cluster.
type ResourceIdentifier struct {
	// Name is the name of the resource as set in metadata.name
	Name string `yaml:"name,omitempty"`
	// Namespace is the namespace of the resource as set in metadata.namespace
	Namespace string `yaml:"namespace,omitempty"`
	// ApiVersion is the apiVersion of the resource
	APIVersion string `yaml:"apiVersion,omitempty"`
	// Kind is the kind of the resource
	Kind string `yaml:"kind,omitempty"`
}

func (r *ResourceIdentifier) GetName() string {
	return r.Name
}

func (r *ResourceIdentifier) GetNamespace() string {
	return r.Namespace
}

func (r *ResourceIdentifier) GetAPIVersion() string {
	return r.APIVersion
}

func (r *ResourceIdentifier) GetKind() string {
	return r.Kind
}

var ErrMissingMetadata = fmt.Errorf("missing Resource metadata")

// GetMeta returns the ResourceMeta for a RNode
func (rn *RNode) GetMeta() (ResourceMeta, error) {
	m := ResourceMeta{}
	b := &bytes.Buffer{}
	e := NewEncoder(b)
	if err := e.Encode(rn.YNode()); err != nil {
		return m, errors.Wrap(err)
	}
	if err := e.Close(); err != nil {
		return m, errors.Wrap(err)
	}
	d := yaml.NewDecoder(b)
	d.KnownFields(false) // only want to parse the metadata
	if err := d.Decode(&m); err != nil {
		return m, errors.Wrap(err)
	}
	if reflect.DeepEqual(m, ResourceMeta{}) {
		return m, ErrMissingMetadata
	}
	return m, nil
}

// Pipe sequentially invokes each GrepFilter, and passes the result to the next
// GrepFilter.
//
// Analogous to http://www.linfo.org/pipes.html
//
// * rn is provided as input to the first GrepFilter.
// * if any GrepFilter returns an error, immediately return the error
// * if any GrepFilter returns a nil RNode, immediately return nil, nil
// * if all Filters succeed with non-empty results, return the final result
func (rn *RNode) Pipe(functions ...Filter) (*RNode, error) {
	// check if rn is nil to make chaining Pipe calls easier
	if rn == nil {
		return nil, nil
	}

	var v *RNode
	var err error
	if rn.value != nil && rn.value.Kind == yaml.DocumentNode {
		// the first node may be a DocumentNode containing a single MappingNode
		v = &RNode{value: rn.value.Content[0]}
	} else {
		v = rn
	}

	// return each fn in sequence until encountering an error or missing value
	for _, c := range functions {
		v, err = c.Filter(v)
		if err != nil || v == nil {
			return v, errors.Wrap(err)
		}
	}
	return v, err
}

// PipeE runs Pipe, dropping the *RNode return value.
// Useful for directly returning the Pipe error value from functions.
func (rn *RNode) PipeE(functions ...Filter) error {
	_, err := rn.Pipe(functions...)
	return errors.Wrap(err)
}

// Document returns the Node RNode for the value.  Does not unwrap the node if it is a
// DocumentNodes
func (rn *RNode) Document() *yaml.Node {
	return rn.value
}

// YNode returns the yaml.Node value.  If the yaml.Node value is a DocumentNode,
// YNode will return the DocumentNode Content entry instead of the DocumentNode.
func (rn *RNode) YNode() *yaml.Node {
	if rn == nil || rn.value == nil {
		return nil
	}
	if rn.value.Kind == yaml.DocumentNode {
		return rn.value.Content[0]
	}
	return rn.value
}

// SetYNode sets the yaml.Node value on an RNode.
func (rn *RNode) SetYNode(node *yaml.Node) {
	if rn.value == nil || node == nil {
		rn.value = node
		return
	}
	*rn.value = *node
}

// AppendToFieldPath appends a field name to the FieldPath.
func (rn *RNode) AppendToFieldPath(parts ...string) {
	rn.fieldPath = append(rn.fieldPath, parts...)
}

// FieldPath returns the field path from the Resource root node, to rn.
// Does not include list indexes.
func (rn *RNode) FieldPath() []string {
	return rn.fieldPath
}

const (
	Trim = "Trim"
	Flow = "Flow"
)

// String returns a string value for a Node, applying the supplied formatting options
func String(node *yaml.Node, opts ...string) (string, error) {
	if node == nil {
		return "", nil
	}
	optsSet := sets.String{}
	optsSet.Insert(opts...)
	if optsSet.Has(Flow) {
		oldStyle := node.Style
		defer func() {
			node.Style = oldStyle
		}()
		node.Style = yaml.FlowStyle
	}

	b := &bytes.Buffer{}
	e := NewEncoder(b)
	err := e.Encode(node)
	e.Close()
	val := b.String()
	if optsSet.Has(Trim) {
		val = strings.TrimSpace(val)
	}
	return val, errors.Wrap(err)
}

// String returns string representation of the RNode
func (rn *RNode) String() (string, error) {
	if rn == nil {
		return "", nil
	}
	return String(rn.value)
}

// MustString returns string representation of the RNode or panics if there is an error
func (rn *RNode) MustString() string {
	s, err := rn.String()
	if err != nil {
		panic(err)
	}
	return s
}

// Content returns Node Content field.
func (rn *RNode) Content() []*yaml.Node {
	if rn == nil {
		return nil
	}
	return rn.YNode().Content
}

// Fields returns the list of field names for a MappingNode.
// Returns an error for non-MappingNodes.
func (rn *RNode) Fields() ([]string, error) {
	if err := ErrorIfInvalid(rn, yaml.MappingNode); err != nil {
		return nil, errors.Wrap(err)
	}
	var fields []string
	for i := 0; i < len(rn.Content()); i += 2 {
		fields = append(fields, rn.Content()[i].Value)
	}
	return fields, nil
}

// Field returns a fieldName, fieldValue pair for MappingNodes.
// Returns nil for non-MappingNodes.
func (rn *RNode) Field(field string) *MapNode {
	if rn.YNode().Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(rn.Content()); i = IncrementFieldIndex(i) {
		isMatchingField := rn.Content()[i].Value == field
		if isMatchingField {
			return &MapNode{Key: NewRNode(rn.Content()[i]), Value: NewRNode(rn.Content()[i+1])}
		}
	}
	return nil
}

// VisitFields calls fn for each field in the RNode.
// Returns an error for non-MappingNodes.
func (rn *RNode) VisitFields(fn func(node *MapNode) error) error {
	// get the list of srcFieldNames
	srcFieldNames, err := rn.Fields()
	if err != nil {
		return errors.Wrap(err)
	}

	// visit each field
	for _, fieldName := range srcFieldNames {
		if err := fn(rn.Field(fieldName)); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

// Elements returns the list of elements in the RNode.
// Returns an error for non-SequenceNodes.
func (rn *RNode) Elements() ([]*RNode, error) {
	if err := ErrorIfInvalid(rn, yaml.SequenceNode); err != nil {
		return nil, errors.Wrap(err)
	}
	var elements []*RNode
	for i := 0; i < len(rn.Content()); i++ {
		elements = append(elements, NewRNode(rn.Content()[i]))
	}
	return elements, nil
}

// ElementValues returns a list of all observed values for a given field name in a
// list of elements.
// Returns error for non-SequenceNodes.
func (rn *RNode) ElementValues(key string) ([]string, error) {
	if err := ErrorIfInvalid(rn, yaml.SequenceNode); err != nil {
		return nil, errors.Wrap(err)
	}
	var elements []string
	for i := 0; i < len(rn.Content()); i++ {
		field := NewRNode(rn.Content()[i]).Field(key)
		if !IsFieldEmpty(field) {
			elements = append(elements, field.Value.YNode().Value)
		}
	}
	return elements, nil
}

// Element returns the element in the list which contains the field matching the value.
// Returns nil for non-SequenceNodes or if no Element matches.
func (rn *RNode) Element(key, value string) *RNode {
	if rn.YNode().Kind != yaml.SequenceNode {
		return nil
	}
	elem, err := rn.Pipe(MatchElement(key, value))
	if err != nil {
		return nil
	}
	return elem
}

// VisitElements calls fn for each element in a SequenceNode.
// Returns an error for non-SequenceNodes
func (rn *RNode) VisitElements(fn func(node *RNode) error) error {
	elements, err := rn.Elements()
	if err != nil {
		return errors.Wrap(err)
	}

	for i := range elements {
		if err := fn(elements[i]); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

// AssociativeSequenceKeys is a map of paths to sequences that have associative keys.
// The order sets the precedence of the merge keys -- if multiple keys are present
// in Resources in a list, then the FIRST key which ALL elements in the list have is used as the
// associative key for merging that list.
var AssociativeSequenceKeys = []string{
	"mountPath", "devicePath", "ip", "type", "topologyKey", "name", "containerPort",
}

// IsAssociative returns true if all elements in the list contain an AssociativeSequenceKey
// as a field.
func IsAssociative(nodes []*RNode) bool {
	for i := range nodes {
		node := nodes[i]
		if IsEmpty(node) {
			continue
		}
		if node.IsAssociative() {
			return true
		}
	}
	return false
}

// IsAssociative returns true if the RNode contains an AssociativeSequenceKey as a field.
func (rn *RNode) IsAssociative() bool {
	return rn.GetAssociativeKey() != ""
}

// GetAssociativeKey returns the AssociativeSequenceKey used to merge the elements in the
// SequenceNode, or "" if the  list is not associative.
func (rn *RNode) GetAssociativeKey() string {
	// look for any associative keys in the first element
	for _, key := range AssociativeSequenceKeys {
		if checkKey(key, rn.Content()) {
			return key
		}
	}

	// element doesn't have an associative keys
	return ""
}

// checkKey returns true if all elems have the key
func checkKey(key string, elems []*Node) bool {
	count := 0
	for i := range elems {
		elem := NewRNode(elems[i])
		if elem.Field(key) != nil {
			count++
		}
	}
	return count == len(elems)
}
