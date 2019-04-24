/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inventory

import (
	"encoding/json"

	"sigs.k8s.io/kustomize/pkg/resid"
)

// A Refs is a map from an Item string to a list of Items
// that it is referred
type Refs map[string][]resid.ItemId

func NewRefs() Refs {
	return Refs{}
}

// Merge merges a Refs into an existing Refs
func (rf Refs) Merge(b Refs) Refs {
	for key, value := range b {
		_, ok := rf[key]
		if ok {
			rf[key] = append(rf[key], value...)
		} else {
			rf[key] = value
		}
	}
	return rf
}

// RemoveIfContains removes the reference relationship
//  a --> b
// from the Refs if it exists
func (rf Refs) RemoveIfContains(a string, b resid.ItemId) {
	refs, ok := rf[a]
	if !ok {
		return
	}
	for i, ref := range refs {
		if ref.String() == b.String() {
			rf[a] = append(refs[:i], refs[i+1:]...)
			break
		}
	}
}

// An Inventory contains current refs
// and previous refs
type Inventory struct {
	Current  Refs `json:"current,omitempty"`
	Previous Refs `json:"previous,omitempty"`
}

// NewInventory returns an Inventory object
func NewInventory() *Inventory {
	return &Inventory{
		Current:  NewRefs(),
		Previous: NewRefs(),
	}
}

// UpdateCurrent updates the Inventory given a
// new current Refs
// The existing Current refs is merged into
// the Previous refs
func (a *Inventory) UpdateCurrent(curref Refs) *Inventory {
	if len(a.Previous) > 0 {
		a.Previous.Merge(a.Current)
	} else {
		a.Previous = a.Current
	}
	a.Current = curref
	return a
}

// Prune returns a list of Items that can be pruned
// as well as updates the Inventory
func (a *Inventory) Prune() []resid.ItemId {
	var results []resid.ItemId
	curref := a.Current

	// Remove references that are already in Current refs
	for item, refs := range curref {
		for _, ref := range refs {
			a.Previous.RemoveIfContains(item, ref)
		}
	}
	// Remove items that are already in Current refs
	for item, refs := range a.Previous {
		if len(refs) == 0 {
			if _, ok := curref[item]; ok {
				delete(a.Previous, item)
			}
		}
	}

	// Remove items from the Previous refs
	// that are not referred by others
	for item, refs := range a.Previous {
		if _, ok := curref[item]; ok {
			continue
		}
		if len(refs) == 0 {
			results = append(results, resid.FromString(item))
			delete(a.Previous, item)
		}
	}

	// Remove items from the Previous refs
	// that are referred only by to be deleted items
	for item, refs := range a.Previous {
		if _, ok := curref[item]; ok {
			delete(a.Previous, item)
			continue
		}

		var newRefs []resid.ItemId
		toDelete := true
		for _, ref := range refs {
			if _, ok := curref[ref.String()]; ok {
				toDelete = false
				newRefs = append(newRefs, ref)
			}
		}
		if toDelete {
			results = append(results, resid.FromString(item))
			delete(a.Previous, item)
		} else {
			a.Previous[item] = newRefs
		}
	}
	return results
}

func (a *Inventory) marshal() ([]byte, error) {
	return json.Marshal(a)
}

func (a *Inventory) unMarshal(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *Inventory) UpdateAnnotations(annot map[string]string) error {
	data, err := a.marshal()
	if err != nil {
		return err
	}
	annot[InventoryAnnotation] = string(data)
	return nil
}

func (a *Inventory) LoadFromAnnotation(annot map[string]string) error {
	value, ok := annot[InventoryAnnotation]
	if ok {
		return a.unMarshal([]byte(value))
	}
	return nil
}
