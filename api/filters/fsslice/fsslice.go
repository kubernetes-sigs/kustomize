// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fsslice

import (
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// SetFn sets a value
type SetFn func(*yaml.RNode) error

// SetScalar returns a SetFn to set a scalar value
func SetScalar(value string) SetFn {
	return func(node *yaml.RNode) error {
		return node.PipeE(yaml.FieldSetter{StringValue: value})
	}
}

// SetEntry returns a SetFn to set an entry in a map
func SetEntry(key, value, tag string) SetFn {
	n := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
		Tag:   tag,
	}
	if tag == yaml.StringTag && yaml.IsYaml1_1NonString(n) {
		n.Style = yaml.DoubleQuotedStyle
	}
	return func(node *yaml.RNode) error {
		return node.PipeE(yaml.FieldSetter{
			Name:  key,
			Value: yaml.NewRNode(n),
		})
	}

}

var _ yaml.Filter = Filter{}

// Filter uses an FsSlice to modify fields on a single object
type Filter struct {
	// FieldSpecList list of FieldSpecs to set
	FsSlice types.FsSlice `yaml:"fsSlice"`

	// SetValue is called on each field that matches one of the FieldSpecs
	SetValue SetFn

	// CreateKind is used to create fields that do not exist
	CreateKind yaml.Kind

	// CreateTag is used to set the tag if encountering a null field
	CreateTag string
}

func (fltr Filter) Filter(obj *yaml.RNode) (*yaml.RNode, error) {
	for i := range fltr.FsSlice {
		// apply this FieldSpec
		// create a new filter for each iteration because they
		// store internal state about the field paths
		_, err := (&fieldSpecFilter{
			FieldSpec:  fltr.FsSlice[i],
			SetValue:   fltr.SetValue,
			CreateKind: fltr.CreateKind,
			CreateTag:  fltr.CreateTag,
		}).Filter(obj)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}
