// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator

import (
	"fmt"
	"log"
	"strings"

	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
	"sigs.k8s.io/kustomize/pkg/types"
)

// ResAccumulator accumulates resources and the rules
// used to customize those resources.  It's a ResMap
// plus stuff needed to modify the ResMap.
type ResAccumulator struct {
	resMap  resmap.ResMap
	tConfig *config.TransformerConfig
	varSet  types.VarSet
}

func MakeEmptyAccumulator() *ResAccumulator {
	ra := &ResAccumulator{}
	ra.resMap = resmap.New()
	ra.tConfig = &config.TransformerConfig{}
	ra.varSet = types.VarSet{}
	return ra
}

// ResMap returns a copy of the internal resMap.
func (ra *ResAccumulator) ResMap() resmap.ResMap {
	return ra.resMap.ShallowCopy()
}

// Vars returns a copy of underlying vars.
func (ra *ResAccumulator) Vars() []types.Var {
	return ra.varSet.Set()
}

func (ra *ResAccumulator) AppendAll(
	resources resmap.ResMap) error {
	return ra.resMap.AppendAll(resources)
}

func (ra *ResAccumulator) AbsorbAll(
	resources resmap.ResMap) error {
	return ra.resMap.AbsorbAll(resources)
}

func (ra *ResAccumulator) MergeConfig(
	tConfig *config.TransformerConfig) (err error) {
	ra.tConfig, err = ra.tConfig.Merge(tConfig)
	return err
}

func (ra *ResAccumulator) GetTransformerConfig() *config.TransformerConfig {
	return ra.tConfig
}

func (ra *ResAccumulator) MergeVars(incoming []types.Var) error {
	return ra.varSet.MergeSlice(incoming)
}

func (ra *ResAccumulator) MergeAccumulator(other *ResAccumulator) (err error) {
	err = ra.AppendAll(other.resMap)
	if err != nil {
		return err
	}
	err = ra.MergeConfig(other.tConfig)
	if err != nil {
		return err
	}
	return ra.varSet.MergeSet(&other.varSet)
}

// makeVarReplacementMap returns a map of Var names to
// their final values. The values are strings intended
// for substitution wherever the $(var.Name) occurs.
func (ra *ResAccumulator) makeVarReplacementMap() (map[string]string, error) {
	result := map[string]string{}
	for _, v := range ra.Vars() {
		matched := ra.resMap.GetMatchingIds(
			resid.NewResId(v.ObjRef.GVK(), v.ObjRef.Name).GvknEquals)
		if len(matched) > 1 {
			return nil, fmt.Errorf(
				"found %d resId matches for var %s "+
					"(unable to disambiguate)",
				len(matched), v)
		}
		if len(matched) == 1 {
			s, err := ra.resMap.GetById(matched[0]).GetFieldValue(v.FieldRef.FieldPath)
			if err != nil {
				return nil, fmt.Errorf(
					"field specified in var '%v' "+
						"not found in corresponding resource", v)
			}
			result[v.Name] = s
		} else {
			return nil, fmt.Errorf(
				"var '%v' cannot be mapped to a field "+
					"in the set of known resources", v)
		}
	}
	return result, nil
}

func (ra *ResAccumulator) Transform(t transformers.Transformer) error {
	return t.Transform(ra.resMap)
}

func (ra *ResAccumulator) ResolveVars() error {
	replacementMap, err := ra.makeVarReplacementMap()
	if err != nil {
		return err
	}
	if len(replacementMap) == 0 {
		return nil
	}
	t := transformers.NewRefVarTransformer(
		replacementMap, ra.tConfig.VarReference)
	err = ra.Transform(t)
	if len(t.UnusedVars()) > 0 {
		log.Printf(
			"well-defined vars that were never replaced: %s\n",
			strings.Join(t.UnusedVars(), ","))
	}
	return err
}

func (ra *ResAccumulator) FixBackReferences() (err error) {
	if ra.tConfig.NameReference == nil {
		return nil
	}
	return ra.Transform(transformers.NewNameReferenceTransformer(
		ra.tConfig.NameReference))
}
