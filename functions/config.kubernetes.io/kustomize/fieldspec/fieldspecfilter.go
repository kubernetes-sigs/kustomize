// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// FieldSpecFilter sets a value using a FieldSpec
type FieldSpecFilter struct {
	// FieldSpec contains the path to the value to set.
	FieldSpec FieldSpec `yaml:"fieldSpec"`

	SetValue func(*yaml.RNode) error

	CreateKind yaml.Kind

	path []string
}

func (fsf *FieldSpecFilter) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	if match, err := fsf.FieldSpec.IsMatch(rn); !match || err != nil {
		return rn, err
	}
	fsf.path = strings.Split(fsf.FieldSpec.Path, "/")
	val, err := fsf.filter(rn)
	if err != nil {
		return nil, err
	}
	return val, err
}

func (fsf *FieldSpecFilter) filter(rn *yaml.RNode) (*yaml.RNode, error) {
	if len(fsf.path) == 0 {
		// found the field -- set its value
		return fsf.doSet(rn)
	}
	if rn.YNode().Kind == yaml.SequenceNode {
		// iterate over all sequence elements
		return fsf.doSeq(rn)
	}

	// recurse on the next path element
	return fsf.doField(rn)
}

func (fsf *FieldSpecFilter) doSet(rn *yaml.RNode) (*yaml.RNode, error) {
	return rn, fsf.SetValue(rn)
}

// doField applies the value to the field matching the next path element.
// may create the field if it is not present
func (fsf *FieldSpecFilter) doField(rn *yaml.RNode) (*yaml.RNode, error) {
	// if the field is a sequence node, it will have a '[]' suffix.
	// strip the suffix so we can lookup the field
	isSeq := strings.HasSuffix(fsf.path[0], "[]")
	fsf.path[0] = strings.TrimSuffix(fsf.path[0], "[]")

	matcher := yaml.FieldMatcher{Name: fsf.path[0]}

	// never create sequence nodes, they would just be empty and we wouldn't recurse into them
	create := fsf.FieldSpec.CreateIfNotPresent && !isSeq && fsf.CreateKind != 0
	if create {
		matcher.Create = yaml.NewRNode(&yaml.Node{Kind: fsf.CreateKind})
	}
	field, err := rn.Pipe(matcher)
	if err != nil || field == nil {
		return field, err
	}

	// recurse on the field, removing the current path element
	recurse := FieldSpecFilter{
		FieldSpec:  fsf.FieldSpec,
		SetValue:   fsf.SetValue,
		CreateKind: fsf.CreateKind,
		path:       fsf.path[1:], // pop 1 off the path
	}
	if _, err := recurse.filter(field); err != nil {
		return nil, err
	}
	return rn, nil
}

// doSeq applies the value to all sequence items
func (fsf *FieldSpecFilter) doSeq(rn *yaml.RNode) (*yaml.RNode, error) {
	if err := rn.VisitElements(func(node *yaml.RNode) error {
		// recurse on each element -- re-allocating a FieldSpecFilter is
		// not strictly required, but is more consistent with doField
		// and less likely to have side effects
		recurse := FieldSpecFilter{
			FieldSpec:  fsf.FieldSpec,
			SetValue:   fsf.SetValue,
			CreateKind: fsf.CreateKind,
			path:       fsf.path, // keep the entire path -- it does not contain parts for sequences
		}
		_, err := recurse.filter(node)
		return err
	}); err != nil {
		return nil, err
	}

	return rn, nil
}
