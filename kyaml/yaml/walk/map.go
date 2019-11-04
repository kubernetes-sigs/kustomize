// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"sort"

	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// walkMap returns the value of VisitMap
//
// - call VisitMap
// - set the return value on l.Dest
// - walk each source field
// - set each source field value on l.Dest
func (l Walker) walkMap() (*yaml.RNode, error) {
	// get the new map value
	dest, err := l.Sources.setDestNode(l.VisitMap(l.Sources))
	if dest == nil || err != nil {
		return nil, err
	}

	// recursively set the field values on the map
	for _, key := range l.fieldNames() {
		val, err := Walker{Visitor: l,
			Sources: l.fieldValue(key), Path: append(l.Path, key)}.Walk()
		if err != nil {
			return nil, err
		}

		// this handles empty and non-empty values
		_, err = dest.Pipe(yaml.FieldSetter{Name: key, Value: val})
		if err != nil {
			return nil, err
		}
	}

	return dest, nil
}

// valueIfPresent returns node.Value if node is non-nil, otherwise returns nil
func (l Walker) valueIfPresent(node *yaml.MapNode) *yaml.RNode {
	if node == nil {
		return nil
	}
	return node.Value
}

// fieldNames returns a sorted slice containing the names of all fields that appear in any of
// the sources
func (l Walker) fieldNames() []string {
	fields := sets.String{}
	for _, s := range l.Sources {
		if s == nil {
			continue
		}
		// don't check error, we know this is a mapping node
		sFields, _ := s.Fields()
		fields.Insert(sFields...)
	}
	result := fields.List()
	sort.Strings(result)
	return result
}

// fieldValue returns a slice containing each source's value for fieldName
func (l Walker) fieldValue(fieldName string) []*yaml.RNode {
	var fields []*yaml.RNode
	for i := range l.Sources {
		if l.Sources[i] == nil {
			fields = append(fields, nil)
			continue
		}
		field := l.Sources[i].Field(fieldName)
		fields = append(fields, l.valueIfPresent(field))
	}
	return fields
}
