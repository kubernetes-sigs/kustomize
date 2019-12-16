// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type FieldSpec struct {
	Gvk `json:",inline,omitempty" yaml:",inline,omitempty"`

	// TODO: make this a []string
	Path               string `json:"path,omitempty" yaml:"path,omitempty"`
	CreateIfNotPresent bool   `json:"create,omitempty" yaml:"create,omitempty"`
}

type FieldSpecList struct {
	Items []FieldSpec `json:"items,omitempty" yaml:"items,omitempty"`
}

type Gvk struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

func (fs FieldSpec) IsMatch(on *yaml.RNode) (bool, error) {
	meta, err := on.GetMeta()
	if err != nil {
		return false, err
	}
	if fs.Kind != "" && fs.Kind != meta.Kind {
		// wrong kind
		return false, err
	}
	var group, version string
	parts := strings.Split(meta.APIVersion, "/")
	if len(parts) > 1 {
		group = parts[0]
		version = parts[1]
	} else {
		version = parts[0]
		group = "core"
	}
	if fs.Group != "" && fs.Group != group {
		// wrong kind
		return false, nil
	}
	if fs.Version != "" && fs.Version != version {
		// wrong kind
		return false, nil
	}

	return true, nil
}

// FieldSpecListFilter applies a Value to a FieldSpecList
type FieldSpecListFilter struct {
	// FieldSpecList list of FieldSpecs to set
	FieldSpecList FieldSpecList `yaml:"fieldSpecList"`

	SetValue func(*yaml.RNode) error

	CreateKind yaml.Kind
}

func (fslf *FieldSpecListFilter) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	for i := range fslf.FieldSpecList.Items {
		f := &FieldSpecFilter{
			FieldSpec:  fslf.FieldSpecList.Items[i],
			SetValue:   fslf.SetValue,
			CreateKind: fslf.CreateKind,
		}
		if _, err := f.Filter(rn); err != nil {
			return nil, err
		}
	}
	return rn, nil
}
