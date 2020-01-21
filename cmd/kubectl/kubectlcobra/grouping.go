// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

const (
	GroupingLabel = "kustomize.config.k8s.io/inventory-id"
	GroupingHash  = "kustomize.config.k8s.io/inventory-hash"
)

// retrieveGroupingLabel returns the string value of the GroupingLabel
// for the passed object. Returns error if the passed object is nil or
// is not a grouping object.
func retrieveGroupingLabel(obj runtime.Object) (string, error) {
	var groupingLabel string
	if obj == nil {
		return "", fmt.Errorf("Grouping object is nil.\n")
	}
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return "", err
	}
	labels := accessor.GetLabels()
	groupingLabel, exists := labels[GroupingLabel]
	if !exists {
		return "", fmt.Errorf("Grouping label does not exist for grouping object: %s\n", GroupingLabel)
	}
	return strings.TrimSpace(groupingLabel), nil
}

// isGroupingObject returns true if the passed object has the
// grouping label.
// TODO(seans3): Check type is ConfigMap.
func isGroupingObject(obj runtime.Object) bool {
	if obj == nil {
		return false
	}
	groupingLabel, err := retrieveGroupingLabel(obj)
	if err == nil && len(groupingLabel) > 0 {
		return true
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
	// as metadata for each object, or if it's the grouping object, store it.
	var groupingInfo *resource.Info
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
			groupingInfo = info
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

	if len(inventoryMap) > 0 {
		// Adds the inventory map to the ConfigMap "data" section.
		err := unstructured.SetNestedStringMap(groupingObj.UnstructuredContent(),
			inventoryMap, "data")
		if err != nil {
			return err
		}
		// Adds the hash of the inventory strings as an annotation to the
		// grouping object. Inventory strings must be sorted to make hash
		// deterministic.
		inventoryList := mapKeysToSlice(inventoryMap)
		sort.Strings(inventoryList)
		invHash, err := calcInventoryHash(inventoryList)
		if err != nil {
			return err
		}
		// Add the hash as a suffix to the grouping object's name.
		invHashStr := strconv.FormatUint(uint64(invHash), 16)
		if err := addSuffixToName(groupingInfo, invHashStr); err != nil {
			return err
		}
		annotations := groupingObj.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[GroupingHash] = invHashStr
		groupingObj.SetAnnotations(annotations)
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

// calcInventoryHash returns an unsigned int32 representing the hash
// of the inventory strings. If there is an error writing bytes to
// the hash, then the error is returned; nil is returned otherwise.
// Used to quickly identify the set of resources in the grouping object.
func calcInventoryHash(inv []string) (uint32, error) {
	h := fnv.New32a()
	for _, is := range inv {
		_, err := h.Write([]byte(is))
		if err != nil {
			return uint32(0), err
		}
	}
	return h.Sum32(), nil
}

// retrieveInventoryHash takes a grouping object (encapsulated by
// a resource.Info), and returns the string representing the hash
// of the grouping inventory; returns empty string if the grouping
// object is not in Unstructured format, or if the hash annotation
// does not exist.
func retrieveInventoryHash(groupingInfo *resource.Info) string {
	var invHash = ""
	groupingObj, ok := groupingInfo.Object.(*unstructured.Unstructured)
	if ok {
		annotations := groupingObj.GetAnnotations()
		if annotations != nil {
			invHash = annotations[GroupingHash]
		}
	}
	return invHash
}

// mapKeysToSlice returns the map keys as a slice of strings.
func mapKeysToSlice(m map[string]string) []string {
	s := make([]string, len(m))
	i := 0
	for k := range m {
		s[i] = k
		i++
	}
	return s
}

// addSuffixToName adds the passed suffix (usually a hash) as a suffix
// to the name of the passed object stored in the Info struct. Returns
// an error if the object is not "*unstructured.Unstructured" or if the
// name stored in the object differs from the name in the Info struct.
func addSuffixToName(info *resource.Info, suffix string) error {

	if info == nil {
		return fmt.Errorf("Nil resource.Info")
	}
	suffix = strings.TrimSpace(suffix)
	if len(suffix) == 0 {
		return fmt.Errorf("Passed empty suffix")
	}

	accessor, _ := meta.Accessor(info.Object)
	name := accessor.GetName()
	if name != info.Name {
		return fmt.Errorf("Grouping object (%s) and resource.Info (%s) have different names\n", name, info.Name)
	}
	// Error if name alread has suffix.
	suffix = "-" + suffix
	if strings.HasSuffix(name, suffix) {
		return fmt.Errorf("Name already has suffix: %s\n", name)
	}
	name += suffix
	accessor.SetName(name)
	info.Name = name

	return nil
}
