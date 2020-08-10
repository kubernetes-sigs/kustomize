// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fixsetters

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// FieldMeta contains metadata that may be attached to fields as comments
type FieldMetaV1 struct {
	Schema spec.Schema

	Extensions XKustomize
}

type XKustomize struct {
	SetBy               string               `yaml:"setBy,omitempty" json:"setBy,omitempty"`
	PartialFieldSetters []PartialFieldSetter `yaml:"partialSetters,omitempty" json:"partialSetters,omitempty"`
	FieldSetter         *PartialFieldSetter  `yaml:"setter,omitempty" json:"setter,omitempty"`
}

// PartialFieldSetter defines how to set part of a field rather than the full field
// value.  e.g. the tag part of an image field
type PartialFieldSetter struct {
	// Name is the name of this setter.
	Name string `yaml:"name" json:"name"`

	// Value is the current value that has been set.
	Value string `yaml:"value" json:"value"`
}

// UpgradeV1SetterComment reads the FieldMeta from a node and upgrade the
// setters comment to latest
func (fm *FieldMetaV1) UpgradeV1SetterComment(n *yaml.RNode) error {
	// check for metadata on head and line comments
	comments := []string{n.YNode().LineComment, n.YNode().HeadComment}
	for _, c := range comments {
		if c == "" {
			continue
		}
		c := strings.TrimLeft(c, "#")

		if err := fm.Schema.UnmarshalJSON([]byte(c)); err != nil {
			// note: don't return an error if the comment isn't a fieldmeta struct
			return nil
		}

		fe := fm.Schema.VendorExtensible.Extensions["x-kustomize"]
		if fe == nil {
			return nil
		}
		b, err := json.Marshal(fe)
		if err != nil {
			return errors.Wrap(err)
		}
		// delete line comment after parsing info into fieldmeta
		n.YNode().HeadComment = ""
		n.YNode().LineComment = ""
		err = json.Unmarshal(b, &fm.Extensions)
		if fm.Extensions.FieldSetter != nil {
			n.YNode().LineComment = fmt.Sprintf(`{"%s":"%s"}`, fieldmeta.ShortHandRef(), fm.Extensions.FieldSetter.Name)
		}
		return err
	}
	return nil
}
