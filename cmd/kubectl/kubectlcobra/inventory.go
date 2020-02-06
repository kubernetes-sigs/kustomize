// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Separates inventory fields. This string is allowable as a
// ConfigMap key, but it is not allowed as a character in
// resource name.
const fieldSeparator = "_"

// Inventory organizes and stores the indentifying information
// for an object. This struct (as a string) is stored in a
// grouping object to keep track of sets of applied objects.
type Inventory struct {
	Namespace string
	Name      string
	GroupKind schema.GroupKind
}

// createInventory returns a pointer to an Inventory struct filled
// with the passed values. This function validates the passed fields
// and returns an error for bad parameters.
func createInventory(namespace string,
	name string, gk schema.GroupKind) (*Inventory, error) {

	// Namespace can be empty, but name cannot.
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("Empty name for inventory object")
	}
	if gk.Empty() {
		return nil, fmt.Errorf("Empty GroupKind for inventory object")
	}

	return &Inventory{
		Namespace: strings.TrimSpace(namespace),
		Name:      name,
		GroupKind: gk,
	}, nil
}

// parseInventory takes a string, splits it into its five fields,
// and returns a pointer to an Inventory struct storing the
// five fields. Example inventory string:
//
//   test-namespace/test-name/apps/v1/ReplicaSet
//
// Returns an error if unable to parse and create the Inventory
// struct.
func parseInventory(inv string) (*Inventory, error) {
	parts := strings.Split(inv, fieldSeparator)
	if len(parts) == 4 {
		gk := schema.GroupKind{
			Group: strings.TrimSpace(parts[2]),
			Kind:  strings.TrimSpace(parts[3]),
		}
		return createInventory(parts[0], parts[1], gk)
	}
	return nil, fmt.Errorf("Unable to decode inventory: %s\n", inv)
}

// Equals returns true if the Inventory structs are identical;
// false otherwise.
func (i *Inventory) Equals(other *Inventory) bool {
	if other == nil {
		return false
	}
	return i.String() == other.String()
}

// String create a string version of the Inventory struct.
func (i *Inventory) String() string {
	return fmt.Sprintf("%s%s%s%s%s%s%s",
		i.Namespace, fieldSeparator,
		i.Name, fieldSeparator,
		i.GroupKind.Group, fieldSeparator,
		i.GroupKind.Kind)
}

// InventorySet encapsulates a grouping of unique Inventory
// structs. Organizes the Inventory structs with a map,
// which ensures there are no duplicates. Allows set
// operations such as merging sets and subtracting sets.
type InventorySet struct {
	set map[string]*Inventory
}

// NewInventorySet returns a pointer to an InventorySet
// struct grouping the passed Inventory items.
func NewInventorySet(items []*Inventory) *InventorySet {
	invSet := InventorySet{set: map[string]*Inventory{}}
	invSet.AddItems(items)
	return &invSet
}

// GetItems returns the set of pointers to Inventory
// structs.
func (is *InventorySet) GetItems() []*Inventory {
	items := []*Inventory{}
	for _, item := range is.set {
		items = append(items, item)
	}
	return items
}

// AddItems adds Inventory structs to the set which
// are not already in the set.
func (is *InventorySet) AddItems(items []*Inventory) {
	for _, item := range items {
		if item != nil {
			is.set[item.String()] = item
		}
	}
}

// DeleteItem removes an Inventory struct from the
// set if it exists in the set. Returns true if the
// Inventory item was deleted, false if it did not exist
// in the set.
func (is *InventorySet) DeleteItem(item *Inventory) bool {
	if item == nil {
		return false
	}
	if _, ok := is.set[item.String()]; ok {
		delete(is.set, item.String())
		return true
	}
	return false
}

// Merge combines the unique set of Inventory items from the
// current set with the passed "other" set, returning a new
// set or error. Returns an error if the passed set to merge
// is nil.
func (is *InventorySet) Merge(other *InventorySet) (*InventorySet, error) {
	if other == nil {
		return nil, fmt.Errorf("InventorySet to merge is nil.")
	}
	// Copy the current InventorySet into result
	result := NewInventorySet(is.GetItems())
	result.AddItems(other.GetItems())
	return result, nil
}

// Subtract removes the Inventory items in the "other" set from the
// current set, returning a new set. This does not modify the current
// set. Returns an error if the passed set to subtract is nil.
func (is *InventorySet) Subtract(other *InventorySet) (*InventorySet, error) {
	if other == nil {
		return nil, fmt.Errorf("InventorySet to subtract is nil.")
	}
	// Copy the current InventorySet into result
	result := NewInventorySet(is.GetItems())
	// Remove each item in "other" which exists in "result"
	for _, item := range other.GetItems() {
		result.DeleteItem(item)
	}
	return result, nil
}

// Equals returns true if the "other" inventory set is the same
// as this current inventory set. Relies on the fact that the
// inventory items are sorted for the String() function.
func (is *InventorySet) Equals(other *InventorySet) bool {
	if other == nil {
		return false
	}
	return is.String() == other.String()
}

// String returns a string describing set of Inventory structs.
func (is *InventorySet) String() string {
	strs := []string{}
	for _, item := range is.GetItems() {
		strs = append(strs, item.String())
	}
	sort.Strings(strs)
	return strings.Join(strs, ", ")
}

// Size returns the number of Inventory structs in the set.
func (is *InventorySet) Size() int {
	return len(is.set)
}
