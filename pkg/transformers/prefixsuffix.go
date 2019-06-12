// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package transformers

import (
	"errors"
	"fmt"

	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

// prefixSuffixTransformer contains the prefix, suffix, and the FieldSpecs
// for each field needing a prefix and suffix.
type prefixSuffixTransformer struct {
	prefix           string
	suffix           string
	fieldSpecsToUse  []config.FieldSpec
	fieldSpecsToSkip []config.FieldSpec
}

var _ Transformer = &prefixSuffixTransformer{}

// Not placed in a file yet due to lack of demand.
var prefixSuffixFieldSpecsToSkip = []config.FieldSpec{
	{
		Gvk: gvk.Gvk{Kind: "CustomResourceDefinition"},
	},
}

// NewPrefixSuffixTransformer makes a prefixSuffixTransformer.
func NewPrefixSuffixTransformer(
	np, ns string, fieldSpecs []config.FieldSpec) (Transformer, error) {
	if len(np) == 0 && len(ns) == 0 {
		return NewNoOpTransformer(), nil
	}
	if fieldSpecs == nil {
		return nil, errors.New("fieldSpecs is not expected to be nil")
	}
	return &prefixSuffixTransformer{
		prefix:           np,
		suffix:           ns,
		fieldSpecsToUse:  fieldSpecs,
		fieldSpecsToSkip: prefixSuffixFieldSpecsToSkip}, nil
}

// Transform prepends the prefix and appends the suffix to the field contents.
// TODO: this transformer breaks internal
// ordering and depends on Id hackery.  Rewrite completely.
func (o *prefixSuffixTransformer) Transform(m resmap.ResMap) error {
	// Fill map "mf" with entries subject to name modification, and
	// delete these entries from "m", so that for now m retains only
	// the entries whose names will not be modified.
	mf := resmap.New()
	for id, r := range m.AsMap() {
		found := false
		for _, path := range o.fieldSpecsToSkip {
			if id.Gvk().IsSelected(&path.Gvk) {
				found = true
				break
			}
		}
		if !found {
			mf.AppendWithId(id, r)
			m.Remove(id)
		}
	}

	for id, r := range mf.AsMap() {
		objMap := r.Map()
		for _, path := range o.fieldSpecsToUse {
			if !id.Gvk().IsSelected(&path.Gvk) {
				continue
			}
			err := mutateField(
				objMap,
				path.PathSlice(),
				path.CreateIfNotPresent,
				o.addPrefixSuffix)
			if err != nil {
				return err
			}
			newId := id.CopyWithNewPrefixSuffix(o.prefix, o.suffix)
			m.AppendWithId(newId, r)
		}
	}
	return nil
}

func (o *prefixSuffixTransformer) addPrefixSuffix(
	in interface{}) (interface{}, error) {
	s, ok := in.(string)
	if !ok {
		return nil, fmt.Errorf("%#v is expected to be %T", in, s)
	}
	return fmt.Sprintf("%s%s%s", o.prefix, s, o.suffix), nil
}
