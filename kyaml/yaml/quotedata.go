// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"strings"
)

// stringDataFields returns the string-only map fields for kind, or nil.
// ConfigMap and Secret data maps cannot legally hold non-string YAML types.
func stringDataFields(kind string) []string {
	switch kind {
	case "ConfigMap":
		return []string{DataField, BinaryDataField}
	case "Secret":
		return []string{DataField, StringDataField}
	default:
		return nil
	}
}

// HasStringDataMaps reports whether rn is a resource whose data values must
// be YAML strings (ConfigMap or Secret).
func HasStringDataMaps(rn *RNode) bool {
	if rn == nil {
		return false
	}
	return stringDataFields(rn.GetKind()) != nil
}

// QuoteStringDataMaps quotes every ConfigMap and Secret string-map value.
// Kubernetes only allows strings in those maps; no other YAML type is legal.
//
// kustomize round-trips resources through JSON before printing YAML, which
// drops original quoting. A value such as ${TEST} is a valid unquoted YAML
// string, but a later envsubst of true or 1 would make the ConfigMap invalid.
// Values that are already quoted are left as-is. Multiline block scalars
// already encode as strings.
func (rn *RNode) QuoteStringDataMaps() (bool, error) {
	if rn == nil {
		return false, nil
	}
	fields := stringDataFields(rn.GetKind())
	if fields == nil {
		return false, nil
	}
	changed := false
	for _, field := range fields {
		didQuote, err := quoteMappingScalars(rn, field)
		if err != nil {
			return false, err
		}
		changed = changed || didQuote
	}
	return changed, nil
}

// QuoteStringDataMapsInYAML parses YAML, quotes ConfigMap and Secret string
// data values, and returns the re-encoded resource. Documents that are not
// ConfigMaps/Secrets, or that need no extra quoting, are returned unchanged
// so existing encoder output is preserved.
func QuoteStringDataMapsInYAML(yml []byte) ([]byte, error) {
	rn, err := Parse(string(yml))
	if err != nil {
		return nil, err
	}
	if !HasStringDataMaps(rn) {
		return yml, nil
	}
	changed, err := rn.QuoteStringDataMaps()
	if err != nil {
		return nil, err
	}
	if !changed {
		return yml, nil
	}
	out, err := rn.String()
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func quoteMappingScalars(rn *RNode, field string) (bool, error) {
	n, err := rn.Pipe(Lookup(field))
	if err != nil {
		return false, err
	}
	if n == nil || n.YNode() == nil || n.YNode().Kind != MappingNode {
		return false, nil
	}
	changed := false
	err = n.VisitFields(func(node *MapNode) error {
		if node == nil || node.Value == nil {
			return nil
		}
		if quoteScalarAsString(node.Value.YNode()) {
			changed = true
		}
		return nil
	})
	return changed, err
}

func quoteScalarAsString(n *Node) bool {
	if n == nil || n.Kind != ScalarNode {
		return false
	}
	if n.Tag == NodeTagNull || IsYNodeTaggedNull(n) {
		return false
	}
	// Block scalars already encode as strings and are a better representation
	// of multiline file contents than a double-quoted string.
	if n.Style&LiteralStyle != 0 || n.Style&FoldedStyle != 0 {
		return false
	}
	if n.Style&DoubleQuotedStyle != 0 || n.Style&SingleQuotedStyle != 0 {
		return false
	}
	if strings.Contains(n.Value, "\n") {
		n.Style = LiteralStyle
		n.Tag = NodeTagString
		return true
	}
	n.Tag = NodeTagString
	n.Style = DoubleQuotedStyle
	return true
}
