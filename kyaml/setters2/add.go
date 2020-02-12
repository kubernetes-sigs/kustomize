// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"strings"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Add creates or updates setter or substitution references from resource fields.
// Requires that at least one of FieldValue and FieldName have been set.
type Add struct {
	// FieldValue if set will add the OpenAPI reference to fields if they have this value.
	// Optional.  If unspecified match all field values.
	FieldValue string

	// FieldName if set will add the OpenAPI reference to fields with this name or path
	// FieldName may be the full name of the field, full path to the field, or the path suffix.
	// e.g. all of the following would match spec.template.spec.containers.image --
	// [image, containers.image, spec.containers.image, template.spec.containers.image,
	//  spec.template.spec.containers.image]
	// Optional.  If unspecified match all field names.
	FieldName string

	// Ref is the OpenAPI reference to set on the matching fields as a comment.
	Ref string
}

// Filter implements yaml.Filter
func (a *Add) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	if a.FieldName == "" && a.FieldValue == "" {
		return nil, errors.Errorf("must specify either fieldName or fieldValue")
	}
	if a.Ref == "" {
		return nil, errors.Errorf("must specify ref")
	}
	return object, accept(a, object)
}

// visitScalar implements visitor
// visitScalar will set the field metadata on each scalar field whose name + value match
func (a *Add) visitScalar(object *yaml.RNode, p string) error {
	// check if the field matches
	if a.FieldName != "" && !strings.HasSuffix(p, a.FieldName) {
		return nil
	}
	if a.FieldValue != "" && a.FieldValue != object.YNode().Value {
		return nil
	}

	// read the field metadata
	fm := fieldmeta.FieldMeta{}
	if err := fm.Read(object); err != nil {
		return err
	}

	// create the ref on the field metadata
	r, err := spec.NewRef(a.Ref)
	if err != nil {
		return err
	}
	fm.Schema.Ref = r

	// write the field metadata
	if err := fm.Write(object); err != nil {
		return err
	}
	return nil
}
