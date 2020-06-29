// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin

import (
	"fmt"
	"strconv"

	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	idAnnotation       = "kustomize.config.k8s.io/id"
	HashAnnotation     = "kustomize.config.k8s.io/needs-hash"
	BehaviorAnnotation = "kustomize.config.k8s.io/behavior"
)

// Returns a new copy of the given ResMap with the ResIds annotated in each Resource
func getResMapWithIdAnnotation(rm resmap.ResMap) (resmap.ResMap, error) {
	inputRM := rm.DeepCopy()
	for _, r := range inputRM.Resources() {
		idString, err := yaml.Marshal(r.CurId())
		if err != nil {
			return nil, err
		}
		annotations := r.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[idAnnotation] = string(idString)
		r.SetAnnotations(annotations)
	}
	return inputRM, nil
}

// updateResMapValues updates the Resource value in the given ResMap
// with the emitted Resource values in output.
func updateResMapValues(pluginName string, h *resmap.PluginHelpers, output []byte, rm resmap.ResMap) error {
	outputRM, err := h.ResmapFactory().NewResMapFromBytes(output)
	if err != nil {
		return err
	}
	for _, r := range outputRM.Resources() {
		// for each emitted Resource, find the matching Resource in the original ResMap
		// using its id
		annotations := r.GetAnnotations()
		idString, ok := annotations[idAnnotation]
		if !ok {
			return fmt.Errorf("the transformer %s should not remove annotation %s",
				pluginName, idAnnotation)
		}
		id := resid.ResId{}
		err := yaml.Unmarshal([]byte(idString), &id)
		if err != nil {
			return err
		}
		res, err := rm.GetByCurrentId(id)
		if err != nil {
			return fmt.Errorf("unable to find unique match to %s", id.String())
		}
		// remove the annotation set by Kustomize to track the resource
		delete(annotations, idAnnotation)
		if len(annotations) == 0 {
			annotations = nil
		}
		r.SetAnnotations(annotations)

		// update the resource value with the transformed object
		res.ResetPrimaryData(r)
	}
	return nil
}

// updateResourceOptions updates the generator options for each resource in the
// given ResMap based on plugin provided annotations.
func UpdateResourceOptions(rm resmap.ResMap) (resmap.ResMap, error) {
	for _, r := range rm.Resources() {
		// Disable name hashing by default and require plugin to explicitly
		// request it for each resource.
		annotations := r.GetAnnotations()
		behavior := annotations[BehaviorAnnotation]
		var needsHash bool
		if val, ok := annotations[HashAnnotation]; ok {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf(
					"the annotation %q contains an invalid value (%q)",
					HashAnnotation, val)
			}
			needsHash = b
		}
		delete(annotations, HashAnnotation)
		delete(annotations, BehaviorAnnotation)
		if len(annotations) == 0 {
			annotations = nil
		}
		r.SetAnnotations(annotations)
		r.SetOptions(types.NewGenArgs(
			&types.GeneratorArgs{
				Behavior: behavior,
				Options:  &types.GeneratorOptions{DisableNameSuffixHash: !needsHash}}))
	}
	return rm, nil
}
