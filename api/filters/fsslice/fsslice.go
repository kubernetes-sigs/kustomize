// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fsslice

import (
	"strings"

	"sigs.k8s.io/kustomize/api/resid"
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
func SetEntry(key, value string) SetFn {
	return func(node *yaml.RNode) error {
		return node.PipeE(yaml.FieldSetter{
			Name: key, StringValue: value})
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
		}).Filter(obj)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

// GetGVK parses the metadata into a GVK
func GetGVK(meta yaml.ResourceMeta) resid.Gvk {
	// parse the group and version from the apiVersion field
	var group, version string
	parts := strings.SplitN(meta.APIVersion, "/", 2)
	group = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}

	return resid.Gvk{
		Group:   group,
		Version: version,
		Kind:    meta.Kind,
	}
}
