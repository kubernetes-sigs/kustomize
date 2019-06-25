// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/transformers"
	"sigs.k8s.io/kustomize/v3/pkg/transformers/config"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

// ResAccumulator accumulates resources and the rules
// used to customize those resources.  It's a ResMap
// plus stuff needed to modify the ResMap.
type ResAccumulator struct {
	resMap         resmap.ResMap
	tConfig        *config.TransformerConfig
	varSet         types.VarSet
	patchSet       []types.Patch
	unresolvedVars types.VarSet
}

func MakeEmptyAccumulator() *ResAccumulator {
	ra := &ResAccumulator{}
	ra.resMap = resmap.New()
	ra.tConfig = &config.TransformerConfig{}
	ra.varSet = types.NewVarSet()
	ra.patchSet = []types.Patch{}
	ra.unresolvedVars = types.NewVarSet()
	return ra
}

// ResMap returns a copy of the internal resMap.
func (ra *ResAccumulator) ResMap() resmap.ResMap {
	return ra.resMap.ShallowCopy()
}

// Vars returns a copy of underlying vars.
func (ra *ResAccumulator) Vars() []types.Var {
	completeset := types.NewVarSet()
	completeset.AbsorbSet(ra.varSet)
	completeset.AbsorbSet(ra.unresolvedVars)
	return completeset.AsSlice()
}

// accumlatePatch accumulates the information regarding conflicting
// resources as patches.
func (ra *ResAccumulator) accumlatePatch(id resid.ResId, conflicting ...*resource.Resource) error {
	target := types.Selector{
		Gvk: gvk.Gvk{
			Group:   id.Group,
			Version: id.Version,
			Kind:    id.Kind,
		},
		Namespace:          id.Namespace,
		Name:               id.Name,
		AnnotationSelector: "",
		LabelSelector:      "",
	}

	for _, res := range conflicting {
		out, err := res.AsYAML()
		if err != nil {
			return err
		}
		newPatch := types.Patch{
			Path:   "",
			Patch:  string(out),
			Target: &target,
		}
		ra.patchSet = append(ra.patchSet, newPatch)
	}
	return nil
}

// HandoverConflictingResources removes conflicting resources from the local accumulator
// and add the conflicting resources list in the other accumulator.
// Conflicting is defined as have the same CurrentId but different Values.
func (ra *ResAccumulator) HandoverConflictingResources(other *ResAccumulator) error {
	for _, rightResource := range other.ResMap().Resources() {
		rightId := rightResource.CurId()
		leftResources := ra.resMap.GetMatchingResourcesByCurrentId(rightId.Equals)

		if len(leftResources) == 0 {
			// no conflict
			continue
		}

		// TODO(jeb): Not sure we want to use DeepEqual here since there are some fields
		// which are artifacts (nameprefix, namesuffix, refvar, refby) added to the resources
		// by the algorithm here.
		// Also we may be dropping here some of the artifacts (nameprefix, namesuffix,...)
		// during the conversion from Resource/ResMap to Patch.
		if len(leftResources) != 1 || !reflect.DeepEqual(leftResources[0], rightResource) {
			// conflict detected. More than one resource or left and right are different.
			if err := other.accumlatePatch(rightId, rightResource); err != nil {
				return err
			}
			if err := other.accumlatePatch(rightId, leftResources...); err != nil {
				return err
			}
		}

		// Remove the resource from that resMap
		err := ra.resMap.Remove(rightId)
		if err != nil {
			return err
		}
	}

	return nil
}

// PatchSet return the list of resources that have been
// put aside. It will let the PatchTransformer decide how to handle
// the conflict, assuming the Transformer can.
func (ra *ResAccumulator) PatchSet() []types.Patch {
	return ra.patchSet
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

// AppendUnresolvedVars accumulates the non conflicting and unresolved variables
// from one accumulator into another. Returns an error if a conflict is detected.
func (ra *ResAccumulator) AppendUnresolvedVars(otherSet types.VarSet) error {
	return ra.unresolvedVars.AbsorbSet(otherSet)
}

// AppendResolvedVars accumulates the non conflicting and resolved variables
// from one accumulator into another. Returns an error is a conflict is detected.
func (ra *ResAccumulator) AppendResolvedVars(otherSet types.VarSet) error {
	mergeableVars := []types.Var{}
	for _, v := range otherSet.AsSlice() {
		conflicting := ra.varSet.Get(v.Name)
		if conflicting == nil {
			// no conflict. The var is valid.
			mergeableVars = append(mergeableVars, v)
			continue
		}

		if !v.DeepEqual(*conflicting) {
			// two vars with the same name are pointing at two
			// different resources.
			return fmt.Errorf(
				"var '%s' already encountered", v.Name)
		}

		// Behavior has been modified.
		// Conflict detection moved to VarMap object
		// =============================================
		// matched := ra.resMap.GetMatchingResourcesByOriginalId(
		// 	resid.NewResId(v.ObjRef.GVK(), v.ObjRef.Name).GvknEquals)
		// if len(matched) > 1 {
		// 	// We detected a diamond import of kustomization
		// 	// context where one variable pointing at one resource
		// 	// in each context is now pointing at two resources
		// 	// (different CurrId) because the two contexts have
		// 	// been merged.
		// 	return fmt.Errorf(
		// 		"found %d resId matches for var %s "+
		// 			"(unable to disambiguate)",
		// 		len(matched), v)
		// }
	}
	return ra.varSet.MergeSlice(mergeableVars)
}

func (ra *ResAccumulator) MergeVars(incoming []types.Var) error {
	// Absorb the new slice of vars into the previously
	// unresolved vars (from the base folders)
	toresolve := ra.unresolvedVars.Copy()
	if err := toresolve.AbsorbSlice(incoming); err != nil {
		return err
	}
	ra.unresolvedVars = types.NewVarSet()

	for _, v := range toresolve.AsSlice() {
		targetId := resid.NewResIdWithNamespace(v.ObjRef.GVK(), v.ObjRef.Name, v.ObjRef.Namespace)
		idMatcher := targetId.GvknEquals
		if targetId.Namespace != "" || !targetId.IsNamespaceableKind() {
			// Preserve backward compatibility. An empty namespace means
			// wildcard search on the namespace hence we still use GvknEquals
			idMatcher = targetId.Equals
		}
		matched := ra.resMap.GetMatchingResourcesByOriginalId(idMatcher)

		// Behavior has been modified.
		// Conflict detection moved to VarMap object
		// =============================================
		// if len(matched) > 1 {
		// 	return fmt.Errorf(
		// 		"found %d resId matches for var %s "+
		// 			"(unable to disambiguate)",
		// 		len(matched), v)
		// }

		if len(matched) == 0 {
			// no associated resources yet.
			if err := ra.unresolvedVars.Absorb(v); err != nil {
				return err
			}
			continue
		}

		// Found one or more associated resource.
		if err := ra.varSet.Absorb(v); err != nil {
			return err
		}
		for _, match := range matched {
			match.AppendRefVarName(v)
		}
	}

	return nil
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
	err = ra.AppendUnresolvedVars(other.unresolvedVars)
	if err != nil {
		return err
	}
	return ra.AppendResolvedVars(other.varSet)
}

func (ra *ResAccumulator) findVarValueFromResources(v types.Var, varMap *resmap.VarMap) error {
	foundCount := 0
	for _, res := range ra.resMap.Resources() {
		for _, varName := range res.GetRefVarNames() {
			if varName == v.Name {
				s, err := res.GetFieldValue(v.FieldRef.FieldPath)
				if err != nil {
					return fmt.Errorf(
						"field specified in var '%v' "+
							"not found in corresponding resource", v)
				}

				foundCount++
				varMap.Append(v.Name, res, s)
			}
		}
	}

	if foundCount == 0 {
		return fmt.Errorf(
			"var '%v' cannot be mapped to a field "+
				"in the set of known resources", v)
	}

	return nil
}

// makeVarReplacementMap returns a map of Var names to
// their final values. The values are strings intended
// for substitution wherever the $(var.Name) occurs.
func (ra *ResAccumulator) makeVarReplacementMap() (resmap.VarMap, error) {
	result := resmap.NewEmptyVarMap()
	for _, v := range ra.Vars() {
		err := ra.findVarValueFromResources(v, &result)
		if err != nil {
			return result, err
		}
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
	if len(replacementMap.VarNames()) == 0 {
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
