// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package resmap implements a map from ResId to Resource that
// tracks all resources in a kustomization.
package resmap

import (
	"fmt"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
)

type varValue struct {
	value interface{}
	ctx   *resource.Resource
}

type VarMap struct {
	varmap map[string][]varValue
}

func NewEmptyVarMap() VarMap {
	return VarMap{varmap: make(map[string][]varValue)}
}

func NewVarMap(simple map[string]interface{}) VarMap {
	varmap := make(map[string][]varValue)
	for key, value := range simple {
		oneVal := varValue{value: value, ctx: nil}
		varmap[key] = []varValue{oneVal}
	}
	return VarMap{varmap: varmap}
}

func (vm *VarMap) Append(name string, ctx *resource.Resource, value interface{}) {
	if _, found := vm.varmap[name]; !found {
		vm.varmap[name] = []varValue{}
	}

	vm.varmap[name] = append(vm.varmap[name], varValue{value: value, ctx: ctx})
}

func (vm *VarMap) VarNames() []string {
	var names []string
	for k := range vm.varmap {
		names = append(names, k)
	}
	return names
}

func (vm *VarMap) SubsetThatCouldBeReferencedByResource(
	inputRes *resource.Resource) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for varName, varValueSlice := range vm.varmap {
		for _, r := range varValueSlice {
			rctxm := inputRes.PrefixesSuffixesEquals
			if (r.ctx == nil) || (r.ctx.InSameKustomizeCtx(rctxm)) {
				if _, found := result[varName]; found {
					return result, fmt.Errorf(
						"found %d resId matches for var %s "+
							"(unable to disambiguate)",
						2, varName)
				}
				result[varName] = r.value
			}
		}
	}
	return result, nil
}
