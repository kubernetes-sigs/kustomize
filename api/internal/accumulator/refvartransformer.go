// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator

import (
	expansion2 "sigs.k8s.io/kustomize/api/internal/accumulator/expansion"

	"sigs.k8s.io/kustomize/api/filters/refvar"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
)

type refVarTransformer struct {
	varMap            map[string]interface{}
	replacementCounts map[string]int
	fieldSpecs        []types.FieldSpec
	mappingFunc       func(string) interface{}
}

// newRefVarTransformer returns a new refVarTransformer
// that replaces $(VAR) style variables with values.
// The fieldSpecs are the places to look for occurrences of $(VAR).
func newRefVarTransformer(
	varMap map[string]interface{}, fs []types.FieldSpec) *refVarTransformer {
	return &refVarTransformer{
		varMap:     varMap,
		fieldSpecs: fs,
	}
}

// UnusedVars returns slice of Var names that were unused
// after a Transform run.
func (rv *refVarTransformer) UnusedVars() []string {
	var unused []string
	for k := range rv.varMap {
		_, ok := rv.replacementCounts[k]
		if !ok {
			unused = append(unused, k)
		}
	}
	return unused
}

// Transform replaces $(VAR) style variables with values.
func (rv *refVarTransformer) Transform(m resmap.ResMap) error {
	rv.replacementCounts = make(map[string]int)
	rv.mappingFunc = expansion2.MappingFuncFor(
		rv.replacementCounts, rv.varMap)
	for _, res := range m.Resources() {
		for _, fieldSpec := range rv.fieldSpecs {
			err := filtersutil.ApplyToJSON(refvar.Filter{
				MappingFunc: rv.mappingFunc,
				FieldSpec:   fieldSpec,
			}, res)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
