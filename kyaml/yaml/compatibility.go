// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"reflect"

	y1_1 "gopkg.in/yaml.v2"
	y1_2 "gopkg.in/yaml.v3"
)

// Style gets the style of the Node
func (rn *RNode) Style() y1_2.Style {
	return rn.YNode().Style
}

// SetStringStyle sets the style for an Node which may be a string in yaml 1.1.
// If it is not a yaml 1.1 string value, the Node style will not be changed.
//
// This is necessary for backwards compatibility with yaml 1.1 from yaml 1.2.
// Kubernetes parses yaml as 1.1 and some values which are strings in  yaml 1.2
// are not parsed as strings in yaml 1.1 (e.g. on is a string in yaml 1.2 and a bool in 1.1).
// These values MUST keep their quoted string style (if the quotes are removed, then Kubernetes will
// parse it as a bool).
func (rn *RNode) SetStringStyle(style y1_2.Style) {
	SetStringStyle(rn.YNode(), style)
}

func SetStringStyle(node *Node, style y1_2.Style) {
	if IsYaml1_1NonString(node) {
		// don't change these styles, they are important for backwards compatibility
		// e.g. "on" must remain quoted, on must remain unquoted
		return
	}
	// style does not have semantic meaning
	node.Style = style
}

// IsYaml1_1NonString returns true if the value parses as a non-string value in yaml 1.1
// when unquoted.
//
// Note: yaml 1.2 uses different keywords than yaml 1.1.  Example: yaml 1.2 interprets
// `field: on` and `field: "on"` as equivalent (both strings).  However Yaml 1.1 interprets
// `field: on` as on being a bool and `field: "on"` as on being a string.
// If an input is read with `field: "on"`, and the style is changed from DoubleQuote to 0,
// it will change the type of the field from a string  to a bool.  For this reason, fields
// which are keywords in yaml 1.1 should never have their style changed, as it would break
// backwards compatibility with yaml 1.1 -- which is what is used by the Kubernetes apiserver.
func IsYaml1_1NonString(node *Node) bool {
	if node.Kind != y1_2.ScalarNode {
		// not a keyword
		return false
	}

	// check if the value unmarshalls differently in yaml 1.1 and yaml 1.2
	var i1 interface{}
	if err := y1_1.Unmarshal([]byte(node.Value), &i1); err != nil {
		// don't touch the style on an unmarshalling error
		return true
	}
	if reflect.TypeOf(i1) != stringType {
		return true
	}

	return false
}

var stringType = reflect.TypeOf("string")
