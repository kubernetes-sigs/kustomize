// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

const GroupingLabel = "kustomize.k8s.io/group-id"

// isGroupingObject returns true if the passed object has the
// grouping label.
// TODO(seans3): Check type is ConfigMap.
func isGroupingObject(obj runtime.Object) bool {
	if obj == nil {
		return false
	}
	accessor, err := meta.Accessor(obj)
	if err == nil {
		labels := accessor.GetLabels()
		_, exists := labels[GroupingLabel]
		if exists {
			return true
		}
	}
	return false
}

// findGroupingObject returns the "Grouping" object (ConfigMap with
// grouping label) if it exists, and a boolean describing if it was found.
func findGroupingObject(infos []*resource.Info) (*resource.Info, bool) {
	for _, info := range infos {
		if info != nil && isGroupingObject(info.Object) {
			return info, true
		}
	}
	return nil, false
}

// sortGroupingObject reorders the infos slice to place the grouping
// object in the first position. Returns true if grouping object found,
// false otherwise.
func sortGroupingObject(infos []*resource.Info) bool {
	for i, info := range infos {
		if info != nil && isGroupingObject(info.Object) {
			// If the grouping object is not already in the first position,
			// swap the grouping object with the first object.
			if i > 0 {
				infos[0], infos[i] = infos[i], infos[0]
			}
			return true
		}
	}
	return false
}

// Adds the inventory of all objects (passed as infos) to the
// grouping object. Returns an error if a grouping object does not
// exist, or we are unable to successfully add the inventory to
// the grouping object; nil otherwise. Each object is in
// unstructured.Unstructured format.
func addInventoryToGroupingObj(infos []*resource.Info) error {

	// Iterate through the objects (infos), creating an Inventory struct
	// as metadata for the object, or if it's the grouping object, store it.
	var groupingObj *unstructured.Unstructured
	inventoryMap := map[string]string{}
	for _, info := range infos {
		obj := info.Object
		if isGroupingObject(obj) {
			// If we have more than one grouping object--error.
			if groupingObj != nil {
				return fmt.Errorf("Error--applying more than one grouping object.")
			}
			var ok bool
			groupingObj, ok = obj.(*unstructured.Unstructured)
			if !ok {
				return fmt.Errorf("Grouping object is not an Unstructured: %#v", groupingObj)
			}
		} else {
			if obj == nil {
				return fmt.Errorf("Creating inventory; object is nil")
			}
			gk := obj.GetObjectKind().GroupVersionKind().GroupKind()
			inventory, err := createInventory(info.Namespace, info.Name, gk)
			if err != nil {
				return err
			}
			inventoryMap[inventory.String()] = ""
		}
	}

	// If we've found the grouping object, store the object metadata inventory
	// in the grouping config map.
	if groupingObj == nil {
		return fmt.Errorf("Grouping object not found")
	}
	err := unstructured.SetNestedStringMap(groupingObj.UnstructuredContent(), inventoryMap, "data")
	if err != nil {
		return err
	}

	return nil
}

// retrieveInventoryFromGroupingObj returns a slice of pointers to the
// inventory metadata. This function finds the grouping object, then
// parses the stored resource metadata into Inventory structs. Returns
// an error if there is a problem parsing the data into Inventory
// structs, or if the grouping object is not in Unstructured format; nil
// otherwise. If a grouping object does not exist, or it does not have a
// "data" map, then returns an empty slice and no error.
func retrieveInventoryFromGroupingObj(infos []*resource.Info) ([]*Inventory, error) {
	inventory := []*Inventory{}
	groupingInfo, exists := findGroupingObject(infos)
	if exists {
		groupingObj, ok := groupingInfo.Object.(*unstructured.Unstructured)
		if !ok {
			err := fmt.Errorf("Grouping object is not an Unstructured: %#v", groupingObj)
			return inventory, err
		}
		invMap, exists, err := unstructured.NestedStringMap(groupingObj.Object, "data")
		if err != nil {
			err := fmt.Errorf("Error retrieving inventory from grouping object.")
			return inventory, err
		}
		if exists {
			for invStr := range invMap {
				inv, err := parseInventory(invStr)
				if err != nil {
					return inventory, err
				}
				inventory = append(inventory, inv)
			}
		}
	}
	return inventory, nil
}
