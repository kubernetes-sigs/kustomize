// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"bytes"
	"errors"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/sets"
)

const (
	// NullNodeTag is the tag set for a yaml.Document that contains no data -- e.g. it isn't a
	// Map, Slice, Document, etc
	NullNodeTag = "!!null"
)

func NullNode() *RNode {
	return NewRNode(&Node{Tag: NullNodeTag})
}

func IsMissingOrNull(node *RNode) bool {
	if node == nil || node.YNode() == nil || node.YNode().Tag == NullNodeTag {
		return true
	}
	return false
}

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

// NewScalarRNode returns a new Scalar *RNode containing the provided value.
func NewScalarRNode(value string) *RNode {
	return &RNode{
		value: &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: value,
		}}
}

// NewListRNode returns a new List *RNode containing the provided value.
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

// NewRNode returns a new *RNode containing the provided value.
func NewRNode(value *yaml.Node) *RNode {
	if value != nil {
		value.Style = 0
	}
	return &RNode{value: value}
}

// GrepFilter may modify or walk the RNode.
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

// ResourceMeta contains the metadata for a Resource.
type ResourceMeta struct {
	ApiVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`
	ObjectMeta `yaml:"metadata,omitempty"`
}

func NewResourceMeta(name string, typeMeta ResourceMeta) ResourceMeta {
	return ResourceMeta{
		Kind:       typeMeta.Kind,
		ApiVersion: typeMeta.ApiVersion,
		ObjectMeta: ObjectMeta{Name: name},
	}
}

type ObjectMeta struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

var MissingMetaError = errors.New("missing Resource metadata")

func (rn *RNode) GetMeta() (ResourceMeta, error) {
	m := ResourceMeta{}
	b := &bytes.Buffer{}
	e := NewEncoder(b)
	if err := e.Encode(rn.YNode()); err != nil {
		return m, err
	}
	if err := e.Close(); err != nil {
		return m, err
	}
	d := yaml.NewDecoder(b)
	d.KnownFields(false) // only want to parse the metadata
	if err := d.Decode(&m); err != nil {
		return m, err
	}
	if reflect.DeepEqual(m, ResourceMeta{}) {
		return m, MissingMetaError
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

	var err error
	var v *RNode
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
			return v, err
		}
	}
	return v, err
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

// SetYNode sets the yaml.Node value.
func (rn *RNode) SetYNode(node *yaml.Node) {
	if rn.value == nil || node == nil {
		rn.value = node
		return
	}
	*rn.value = *node
}

// SetYNode sets the value on a Document.
func (rn *RNode) AppendToFieldPath(parts ...string) {
	rn.fieldPath = append(rn.fieldPath, parts...)
}

// FieldPath returns the field path from the object root to rn.  Does not include list indexes.
func (rn *RNode) FieldPath() []string {
	return rn.fieldPath
}

const (
	Trim = "Trim"
	Flow = "Flow"
)

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
	return val, err
}

// NewScalarRNode returns the yaml NewScalarRNode representation of the RNode value.
func (rn *RNode) String() (string, error) {
	if rn == nil {
		return "", nil
	}
	return String(rn.value)
}

func (rn *RNode) MustString() string {
	s, err := rn.String()
	if err != nil {
		panic(err)
	}
	return s
}

// Content returns the value node's Content field.
func (rn *RNode) Content() []*yaml.Node {
	if rn == nil {
		return nil
	}
	return rn.YNode().Content
}

// Fields returns the list of fields for a ResourceNode containing a MappingNode
// value.
func (rn *RNode) Fields() ([]string, error) {
	if err := ErrorIfInvalid(rn, yaml.MappingNode); err != nil {
		return nil, err
	}
	var fields []string
	for i := 0; i < len(rn.Content()); i += 2 {
		fields = append(fields, rn.Content()[i].Value)
	}
	return fields, nil
}

// Field returns the fieldName, fieldValue pair for MappingNodes.  Returns nil for non-MappingNodes.
func (rn *RNode) Field(field string) *MapNode {
	if rn.YNode().Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(rn.Content()); IncrementFieldIndex(&i) {
		isMatchingField := rn.Content()[i].Value == field
		if isMatchingField {
			return &MapNode{Key: NewRNode(rn.Content()[i]), Value: NewRNode(rn.Content()[i+1])}
		}
	}
	return nil
}

// VisitFields calls fn for each field in rn.
func (rn *RNode) VisitFields(fn func(node *MapNode) error) error {
	// get the list of srcFieldNames
	srcFieldNames, err := rn.Fields()
	if err != nil {
		return err
	}

	// visit each field
	for _, fieldName := range srcFieldNames {
		if err := fn(rn.Field(fieldName)); err != nil {
			return err
		}
	}
	return nil
}

// Elements returns a list of elements for a ResourceNode containing a
// SequenceNode value.
func (rn *RNode) Elements() ([]*RNode, error) {
	if err := ErrorIfInvalid(rn, yaml.SequenceNode); err != nil {
		return nil, err
	}
	var elements []*RNode
	for i := 0; i < len(rn.Content()); i += 1 {
		elements = append(elements, NewRNode(rn.Content()[i]))
	}
	return elements, nil
}

func (rn *RNode) ElementValues(key string) ([]string, error) {
	if err := ErrorIfInvalid(rn, yaml.SequenceNode); err != nil {
		return nil, err
	}
	var elements []string
	for i := 0; i < len(rn.Content()); i += 1 {
		field := NewRNode(rn.Content()[i]).Field(key)
		if !IsFieldEmpty(field) {
			elements = append(elements, field.Value.YNode().Value)
		}
	}
	return elements, nil
}

// Element returns the element in the list which contains the field matching the value.
// Returns nil for non-SequenceNodes
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

// VisitElements calls fn for each element in the list.
func (rn *RNode) VisitElements(fn func(node *RNode) error) error {
	elements, err := rn.Elements()
	if err != nil {
		return err
	}

	for i := range elements {
		if err := fn(elements[i]); err != nil {
			return err
		}
	}
	return nil
}

// AssociativeSequencePaths is a map of paths to sequences that have associative keys.
// The order sets the precedence of the merge keys -- if multiple keys are present
// in the list, then the FIRST key which ALL elements have is used as the
// associative key.
var AssociativeSequenceKeys = []string{
	"mountPath", "devicePath", "ip", "type", "topologyKey", "name", "containerPort",
}

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

// IsAssociative returns true if the RNode is for an associative list.
func (rn *RNode) IsAssociative() bool {
	return rn.GetAssociativeKey() != ""
}

// GetAssociativeKey returns the associative key used to merge the list, or "" if the
// list is not associative.
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
