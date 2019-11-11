// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator

import (
	"fmt"
	"log"
	"strings"

	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
)

// ResAccumulator accumulates resources and the rules
// used to customize those resources.  It's a ResMap
// plus stuff needed to modify the ResMap.
type ResAccumulator struct {
	resMap            resmap.ResMap
	tConfig           *builtinconfig.TransformerConfig
	varSet            types.VarSet
	accumulatedVars   types.VarSet
	unaccumulatedVars types.VarSet
}

func MakeEmptyAccumulator() *ResAccumulator {
	ra := &ResAccumulator{}
	ra.resMap = resmap.New()
	ra.tConfig = &builtinconfig.TransformerConfig{}
	ra.varSet = types.NewVarSet()
	ra.accumulatedVars = types.NewVarSet()
	ra.unaccumulatedVars = types.NewVarSet()
	return ra
}

// ResMap returns a copy of the internal resMap.
func (ra *ResAccumulator) ResMap() resmap.ResMap {
	return ra.resMap.ShallowCopy()
}

// Vars returns a copy of underlying vars.
func (ra *ResAccumulator) Vars() []types.Var {
	return ra.varSet.AsSlice()
}

// AccumulatedVars returns a copy of underlying vars.
func (ra *ResAccumulator) AccumulatedVars() []types.Var {
	return ra.accumulatedVars.AsSlice()
}

// UnaccumulatedVars returns a copy of underlying vars.
func (ra *ResAccumulator) UnaccumulatedVars() []types.Var {
	return ra.unaccumulatedVars.AsSlice()
}

// DeleteVars deletes varset.
func (ra *ResAccumulator) DeleteVars() {
	ra.varSet = types.NewVarSet()
}

// DeleteUnaccumulatedVars deletes unaccumulated varset.
func (ra *ResAccumulator) DeleteUnaccumulatedVars() {
	ra.unaccumulatedVars = types.NewVarSet()
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
	tConfig *builtinconfig.TransformerConfig) (err error) {
	ra.tConfig, err = ra.tConfig.Merge(tConfig)
	return err
}

//MergeVarSet merges the varset through resource accumulators
func (ra *ResAccumulator) MergeVarSet(vars []types.Var) (err error) {
	return ra.varSet.MergeSlice(vars)
}

//MergeUnaccumulatedVarSet merges the varset through resource accumulators
func (ra *ResAccumulator) MergeUnaccumulatedVarSet(vars []types.Var) (err error) {
	return ra.unaccumulatedVars.MergeSlice(vars)
}

//MergeAccumulatedVars merges the accumulatedVars through resource accumulators
func (ra *ResAccumulator) MergeAccumulatedVars(accumulatedVars types.VarSet) (err error) {
	return ra.accumulatedVars.MergeSet(accumulatedVars)
}

func (ra *ResAccumulator) GetTransformerConfig() *builtinconfig.TransformerConfig {
	return ra.tConfig
}

func (ra *ResAccumulator) MergeVars(incoming []types.Var) error {
	for _, v := range incoming {
		targetId := resid.NewResIdWithNamespace(v.ObjRef.GVK(), v.ObjRef.Name, v.ObjRef.Namespace)
		idMatcher := targetId.GvknEquals
		if targetId.Namespace != "" || !targetId.IsNamespaceableKind() {
			// Preserve backward compatibility. An empty namespace means
			// wildcard search on the namespace hence we still use GvknEquals
			idMatcher = targetId.Equals
		}
		matched := ra.resMap.GetMatchingResourcesByOriginalId(idMatcher)
		if len(matched) > 1 {
			return fmt.Errorf(
				"found %d resId matches for var %v "+
					"(unable to disambiguate)",
				len(matched), v)
		}
		if len(matched) == 1 {
			matched[0].AppendRefVarName(v)
			ra.accumulatedVars.Merge(v)
		}
		if len(matched) == 0 {
			ra.unaccumulatedVars.Merge(v)
		}

	}
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
	err = ra.MergeAccumulatedVars(other.accumulatedVars)
	if err != nil {
		return err
	}
	err = ra.MergeUnaccumulatedVarSet(other.UnaccumulatedVars())
	if err != nil {
		return err
	}
	return ra.varSet.MergeSet(other.varSet)
}

func (ra *ResAccumulator) findVarValueFromResources(v types.Var) (interface{}, error) {
	for _, res := range ra.resMap.Resources() {
		for _, varName := range res.GetRefVarNames() {
			if varName == v.Name {
				s, err := res.GetFieldValue(v.FieldRef.FieldPath)
				if err != nil {
					return "", fmt.Errorf(
						"field specified in var '%v' "+
							"not found in corresponding resource", v)
				}

				return s, nil
			}
		}
	}

	return "", fmt.Errorf(
		"var '%v' cannot be mapped to a field "+
			"in the set of known resources", v)
}

// makeVarReplacementMap returns a map of Var names to
// their final values. The values are strings intended
// for substitution wherever the $(var.Name) occurs.
func (ra *ResAccumulator) makeVarReplacementMap() (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for _, v := range ra.Vars() {
		s, err := ra.findVarValueFromResources(v)
		if err != nil {
			return nil, err
		}

		result[v.Name] = s
	}

	return result, nil
}

func (ra *ResAccumulator) Transform(t resmap.Transformer) error {
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
	t := newRefVarTransformer(
		replacementMap, ra.tConfig.VarReference)
	err = ra.Transform(t)
	if len(t.UnusedVars()) > 0 {
		log.Printf(
			"well-defined vars that were never replaced: %s\n",
			strings.Join(t.UnusedVars(), ","))
	}
	return err
}

// if a resource is not found dont error, assume we will find it later.
func (ra *ResAccumulator) makeVarDirectoryReplacementMap() (map[string]interface{}, []types.Var) {
	result := map[string]interface{}{}
	globalVars := []types.Var{}
	for _, v := range ra.AccumulatedVars() {
		if v.IsImmediateSubstitution() {
			s, err := ra.findVarValueFromResources(v)
			if err != nil {
				globalVars = append(globalVars, v)
			}
			result[v.Name] = s
		} else {
			globalVars = append(globalVars, v)
		}
	}

	//add unaccumulatedVars to global
	globalVars = append(globalVars, ra.UnaccumulatedVars()...)

	return result, globalVars
}

// allows for vars to not have a found resource at the time of directory resolution
func (ra *ResAccumulator) ResolveDirectoryVars() error {
	replacementMap, globalVars := ra.makeVarDirectoryReplacementMap()
	varSetCopy := types.NewVarSet()
	accumulatedVarSetCopy := types.NewVarSet()
	if len(replacementMap) == 0 {
		err := varSetCopy.MergeSlice(globalVars)
		ra.varSet = varSetCopy
		return err
	}
	t := newRefVarTransformer(
		replacementMap, ra.tConfig.VarReference)
	err := ra.Transform(t)
	if err != nil {
		return err
	}
	if len(t.UnusedVars()) > 0 {
		for _, i := range t.UnusedVars() {
			unusedVar := ra.varSet.Get(i)
			varSetCopy.Merge(*unusedVar)
			accumulatedVarSetCopy.Merge(*unusedVar)
		}

	}
	ra.accumulatedVars = accumulatedVarSetCopy
	err = varSetCopy.MergeSlice(globalVars)
	ra.varSet = varSetCopy
	return err
}

func (ra *ResAccumulator) FixBackReferences() (err error) {
	if ra.tConfig.NameReference == nil {
		return nil
	}
	return ra.Transform(newNameReferenceTransformer(
		ra.tConfig.NameReference))
}
