// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/sets"
)

const (
	// NodeTagNull is the tag set for a yaml.Document that contains no data;
	// e.g. it isn't a Map, Slice, Document, etc
	NodeTagNull   = "!!null"
	NodeTagFloat  = "!!float"
	NodeTagString = "!!str"
	NodeTagBool   = "!!bool"
	NodeTagInt    = "!!int"
	NodeTagMap    = "!!map"
	NodeTagSeq    = "!!seq"
	NodeTagEmpty  = ""
)

// MakeNullNode returns an RNode that represents an empty document.
func MakeNullNode() *RNode {
	return NewRNode(&Node{Tag: NodeTagNull})
}

// IsMissingOrNull is true if the RNode is nil or explicitly tagged null.
// TODO: make this a method on RNode.
func IsMissingOrNull(node *RNode) bool {
	return node.IsNil() || node.YNode().Tag == NodeTagNull
}

// IsEmptyMap returns true if the RNode is an empty node or an empty map.
// TODO: make this a method on RNode.
func IsEmptyMap(node *RNode) bool {
	return IsMissingOrNull(node) || IsYNodeEmptyMap(node.YNode())
}

// IsYNodeTaggedNull returns true if the node is explicitly tagged Null.
func IsYNodeTaggedNull(n *yaml.Node) bool {
	return n != nil && n.Tag == NodeTagNull
}

// IsYNodeEmptyMap is true if the Node is a non-nil empty map.
func IsYNodeEmptyMap(n *yaml.Node) bool {
	return n != nil && n.Kind == yaml.MappingNode && len(n.Content) == 0
}

// IsYNodeEmptyMap is true if the Node is a non-nil empty sequence.
func IsYNodeEmptySeq(n *yaml.Node) bool {
	return n != nil && n.Kind == yaml.SequenceNode && len(n.Content) == 0
}

// IsYNodeEmptyDoc is true if the node is a Document with no content.
// E.g.: "---\n---"
func IsYNodeEmptyDoc(n *yaml.Node) bool {
	return n.Kind == yaml.DocumentNode && n.Content[0].Tag == NodeTagNull
}

func IsYNodeString(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == NodeTagString
}

// GetValue returns underlying yaml.Node Value field
func GetValue(node *RNode) string {
	if IsMissingOrNull(node) {
		return ""
	}
	return node.YNode().Value
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

// ReadFile parses a single Resource from a yaml file
func ReadFile(path string) (*RNode, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(string(b))
}

// WriteFile writes a single Resource to a yaml file
func WriteFile(node *RNode, path string) error {
	out, err := node.String()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(out), 0600)
}

// UpdateFile reads the file at path, applies the filter to it, and write the result back.
// path must contain a exactly 1 resource (YAML).
func UpdateFile(filter Filter, path string) error {
	// Read the yaml
	y, err := ReadFile(path)
	if err != nil {
		return err
	}

	// Update the yaml
	if err := y.PipeE(filter); err != nil {
		return err
	}

	// Write the yaml
	return WriteFile(y, path)
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

// NewMapRNode returns a new Map *RNode containing the provided values
func NewMapRNode(values *map[string]string) *RNode {
	m := &RNode{value: &yaml.Node{
		Kind: yaml.MappingNode,
	}}
	if values == nil {
		return m
	}

	for k, v := range *values {
		m.value.Content = append(m.value.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: k,
		}, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: v,
		})
	}

	return m
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

// IsNilOrEmpty returns true if the MapNode is nil,
// has no value, or has a value that appears empty.
func (mn *MapNode) IsNilOrEmpty() bool {
	return mn == nil || mn.Value.IsNilOrEmpty()
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

type TypeMeta struct {
	Kind       string
	APIVersion string
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

// Field names
const (
	AnnotationsField = "annotations"
	APIVersionField  = "apiVersion"
	KindField        = "kind"
	MetadataField    = "metadata"
	NameField        = "name"
	NamespaceField   = "namespace"
	LabelsField      = "labels"
)

// IsNil is true if the node is nil, or its underlying YNode is nil.
func (rn *RNode) IsNil() bool {
	return rn == nil || rn.YNode() == nil
}

// IsTaggedNull is true if a non-nil node is explicitly tagged Null.
func (rn *RNode) IsTaggedNull() bool {
	return !rn.IsNil() && IsYNodeTaggedNull(rn.YNode())
}

// IsNilOrEmpty is true if the node is nil,
// has no YNode, or has YNode that appears empty.
func (rn *RNode) IsNilOrEmpty() bool {
	return rn.IsNil() ||
		IsYNodeTaggedNull(rn.YNode()) ||
		IsYNodeEmptyMap(rn.YNode()) ||
		IsYNodeEmptySeq(rn.YNode())
}

// GetMeta returns the ResourceMeta for an RNode
func (rn *RNode) GetMeta() (ResourceMeta, error) {
	if IsMissingOrNull(rn) {
		return ResourceMeta{}, nil
	}
	missingMeta := true
	n := rn
	if n.YNode().Kind == DocumentNode {
		// get the content is this is the document node
		n = NewRNode(n.Content()[0])
	}

	// don't decode into the struct directly or it will fail on UTF-8 issues
	// which appear in comments
	m := ResourceMeta{}

	// TODO: consider optimizing this parsing
	if f := n.Field(APIVersionField); !f.IsNilOrEmpty() {
		m.APIVersion = GetValue(f.Value)
		missingMeta = false
	}
	if f := n.Field(KindField); !f.IsNilOrEmpty() {
		m.Kind = GetValue(f.Value)
		missingMeta = false
	}

	mf := n.Field(MetadataField)
	if mf.IsNilOrEmpty() {
		if missingMeta {
			return m, ErrMissingMetadata
		}
		return m, nil
	}
	meta := mf.Value

	if f := meta.Field(NameField); !f.IsNilOrEmpty() {
		m.Name = f.Value.YNode().Value
		missingMeta = false
	}
	if f := meta.Field(NamespaceField); !f.IsNilOrEmpty() {
		m.Namespace = GetValue(f.Value)
		missingMeta = false
	}

	if f := meta.Field(LabelsField); !f.IsNilOrEmpty() {
		m.Labels = map[string]string{}
		_ = f.Value.VisitFields(func(node *MapNode) error {
			m.Labels[GetValue(node.Key)] = GetValue(node.Value)
			return nil
		})
		missingMeta = false
	}
	if f := meta.Field(AnnotationsField); !f.IsNilOrEmpty() {
		m.Annotations = map[string]string{}
		_ = f.Value.VisitFields(func(node *MapNode) error {
			m.Annotations[GetValue(node.Key)] = GetValue(node.Value)
			return nil
		})
		missingMeta = false
	}

	if missingMeta {
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
		if !field.IsNilOrEmpty() {
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
// Only infer name as a merge key.
var AssociativeSequenceKeys = []string{"name"}

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

func (rn *RNode) MarshalJSON() ([]byte, error) {
	s, err := rn.String()
	if err != nil {
		return nil, err
	}

	if rn.YNode().Kind == SequenceNode {
		var a []interface{}
		if err := Unmarshal([]byte(s), &a); err != nil {
			return nil, err
		}
		return json.Marshal(a)
	}

	m := map[string]interface{}{}
	if err := Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func (rn *RNode) UnmarshalJSON(b []byte) error {
	m := map[string]interface{}{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	c, err := Marshal(m)
	if err != nil {
		return err
	}

	r, err := Parse(string(c))
	if err != nil {
		return err
	}
	rn.value = r.value
	return nil
}

// ConvertJSONToYamlNode parses input json string and returns equivalent yaml node
func ConvertJSONToYamlNode(jsonStr string) (*RNode, error) {
	var body map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &body)
	if err != nil {
		return nil, err
	}
	yml, err := yaml.Marshal(body)
	if err != nil {
		return nil, err
	}
	node, err := Parse(string(yml))
	if err != nil {
		return nil, err
	}
	return node, nil
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
