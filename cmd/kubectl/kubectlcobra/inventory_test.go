// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCreateInventory(t *testing.T) {
	tests := []struct {
		namespace string
		name      string
		gk        schema.GroupKind
		expected  string
		isError   bool
	}{
		{
			namespace: "  \n",
			name:      " test-name\t",
			gk: schema.GroupKind{
				Group: "apps",
				Kind:  "ReplicaSet",
			},
			expected: "_test-name_apps_ReplicaSet",
			isError:  false,
		},
		{
			namespace: "test-namespace ",
			name:      " test-name\t",
			gk: schema.GroupKind{
				Group: "apps",
				Kind:  "ReplicaSet",
			},
			expected: "test-namespace_test-name_apps_ReplicaSet",
			isError:  false,
		},
		// Error with empty name.
		{
			namespace: "test-namespace ",
			name:      " \t",
			gk: schema.GroupKind{
				Group: "apps",
				Kind:  "ReplicaSet",
			},
			expected: "",
			isError:  true,
		},
		// Error with empty GroupKind.
		{
			namespace: "test-namespace",
			name:      "test-name",
			gk:        schema.GroupKind{},
			expected:  "",
			isError:   true,
		},
	}

	for _, test := range tests {
		inv, err := createInventory(test.namespace, test.name, test.gk)
		if !test.isError {
			if err != nil {
				t.Errorf("Error creating inventory when it should have worked.")
			} else if test.expected != inv.String() {
				t.Errorf("Expected inventory (%s) != created inventory(%s)\n", test.expected, inv.String())
			}
		}
		if test.isError && err == nil {
			t.Errorf("Should have returned an error in createInventory()")
		}
	}
}

func TestInventoryEqual(t *testing.T) {
	tests := []struct {
		inventory1 *Inventory
		inventory2 *Inventory
		isEqual    bool
	}{
		// "Other" inventory is nil, then not equal.
		{
			inventory1: &Inventory{
				Name: "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			inventory2: nil,
			isEqual:    false,
		},
		// Two equal inventories without a namespace
		{
			inventory1: &Inventory{
				Name: "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			inventory2: &Inventory{
				Name: "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			isEqual: true,
		},
		// Two equal inventories with a namespace
		{
			inventory1: &Inventory{
				Namespace: "test-namespace",
				Name:      "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			inventory2: &Inventory{
				Namespace: "test-namespace",
				Name:      "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			isEqual: true,
		},
		// One inventory with a namespace, one without -- not equal.
		{
			inventory1: &Inventory{
				Name: "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			inventory2: &Inventory{
				Namespace: "test-namespace",
				Name:      "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			isEqual: false,
		},
		// One inventory with a Deployment, one with a ReplicaSet -- not equal.
		{
			inventory1: &Inventory{
				Name: "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			inventory2: &Inventory{
				Name: "test-inv",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "ReplicaSet",
				},
			},
			isEqual: false,
		},
	}

	for _, test := range tests {
		actual := test.inventory1.Equals(test.inventory2)
		if test.isEqual && !actual {
			t.Errorf("Expected inventories equal, but actual is not: (%s)/(%s)\n", test.inventory1, test.inventory2)
		}
	}
}

func TestParseInventory(t *testing.T) {
	tests := []struct {
		invStr    string
		inventory *Inventory
		isError   bool
	}{
		{
			invStr: "_test-name_apps_ReplicaSet\t",
			inventory: &Inventory{
				Name: "test-name",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "ReplicaSet",
				},
			},
			isError: false,
		},
		{
			invStr: "test-namespace_test-name_apps_Deployment",
			inventory: &Inventory{
				Namespace: "test-namespace",
				Name:      "test-name",
				GroupKind: schema.GroupKind{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			isError: false,
		},
		// Not enough fields -- error
		{
			invStr:    "_test-name_apps",
			inventory: &Inventory{},
			isError:   true,
		},
	}

	for _, test := range tests {
		actual, err := parseInventory(test.invStr)
		if !test.isError {
			if err != nil {
				t.Errorf("Error parsing inventory when it should have worked.")
			} else if !test.inventory.Equals(actual) {
				t.Errorf("Expected inventory (%s) != parsed inventory (%s)\n", test.inventory, actual)
			}
		}
		if test.isError && err == nil {
			t.Errorf("Should have returned an error in parseInventory()")
		}
	}
}

var inventory1 = Inventory{
	Namespace: "test-namespace",
	Name:      "test-inv-1",
	GroupKind: schema.GroupKind{
		Group: "apps",
		Kind:  "Deployment",
	},
}

var inventory2 = Inventory{
	Namespace: "test-namespace",
	Name:      "test-inv-2",
	GroupKind: schema.GroupKind{
		Group: "",
		Kind:  "Pod",
	},
}

var inventory3 = Inventory{
	Namespace: "test-namespace",
	Name:      "test-inv-3",
	GroupKind: schema.GroupKind{
		Group: "",
		Kind:  "Service",
	},
}

var inventory4 = Inventory{
	Namespace: "test-namespace",
	Name:      "test-inv-4",
	GroupKind: schema.GroupKind{
		Group: "apps",
		Kind:  "DaemonSet",
	},
}

func TestNewInventorySet(t *testing.T) {
	tests := []struct {
		items        []*Inventory
		expectedStr  string
		expectedSize int
	}{
		{
			items:        []*Inventory{},
			expectedStr:  "",
			expectedSize: 0,
		},
		{
			items:        []*Inventory{&inventory1},
			expectedStr:  "test-namespace_test-inv-1_apps_Deployment",
			expectedSize: 1,
		},
		{
			items:        []*Inventory{&inventory1, &inventory2},
			expectedStr:  "test-namespace_test-inv-1_apps_Deployment, test-namespace_test-inv-2__Pod",
			expectedSize: 2,
		},
	}

	for _, test := range tests {
		invSet := NewInventorySet(test.items)
		actualStr := invSet.String()
		actualSize := invSet.Size()
		if test.expectedStr != actualStr {
			t.Errorf("Expected InventorySet (%s), got (%s)\n", test.expectedStr, actualStr)
		}
		if test.expectedSize != actualSize {
			t.Errorf("Expected InventorySet size (%d), got (%d)\n", test.expectedSize, actualSize)
		}
		actualItems := invSet.GetItems()
		if len(test.items) != len(actualItems) {
			t.Errorf("Expected num inventory items (%d), got (%d)\n", len(test.items), len(actualItems))
		}
	}
}

func TestInventorySetAddItems(t *testing.T) {
	tests := []struct {
		initialItems  []*Inventory
		addItems      []*Inventory
		expectedItems []*Inventory
	}{
		// Adding no items to empty inventory set.
		{
			initialItems:  []*Inventory{},
			addItems:      []*Inventory{},
			expectedItems: []*Inventory{},
		},
		// Adding item to empty inventory set.
		{
			initialItems:  []*Inventory{},
			addItems:      []*Inventory{&inventory1},
			expectedItems: []*Inventory{&inventory1},
		},
		// Adding no items does not change the inventory set
		{
			initialItems:  []*Inventory{&inventory1},
			addItems:      []*Inventory{},
			expectedItems: []*Inventory{&inventory1},
		},
		// Adding an item which alread exists does not increase size.
		{
			initialItems:  []*Inventory{&inventory1, &inventory2},
			addItems:      []*Inventory{&inventory1},
			expectedItems: []*Inventory{&inventory1, &inventory2},
		},
		{
			initialItems:  []*Inventory{&inventory1, &inventory2},
			addItems:      []*Inventory{&inventory3, &inventory4},
			expectedItems: []*Inventory{&inventory1, &inventory2, &inventory3, &inventory4},
		},
	}

	for _, test := range tests {
		invSet := NewInventorySet(test.initialItems)
		invSet.AddItems(test.addItems)
		if len(test.expectedItems) != invSet.Size() {
			t.Errorf("Expected num inventory items (%d), got (%d)\n", len(test.expectedItems), invSet.Size())
		}
	}
}

func TestInventorySetDeleteItem(t *testing.T) {
	tests := []struct {
		initialItems  []*Inventory
		deleteItem    *Inventory
		expected      bool
		expectedItems []*Inventory
	}{
		{
			initialItems:  []*Inventory{},
			deleteItem:    nil,
			expected:      false,
			expectedItems: []*Inventory{},
		},
		{
			initialItems:  []*Inventory{},
			deleteItem:    &inventory1,
			expected:      false,
			expectedItems: []*Inventory{},
		},
		{
			initialItems:  []*Inventory{&inventory2},
			deleteItem:    &inventory1,
			expected:      false,
			expectedItems: []*Inventory{&inventory2},
		},
		{
			initialItems:  []*Inventory{&inventory1},
			deleteItem:    &inventory1,
			expected:      true,
			expectedItems: []*Inventory{},
		},
		{
			initialItems:  []*Inventory{&inventory1, &inventory2},
			deleteItem:    &inventory1,
			expected:      true,
			expectedItems: []*Inventory{&inventory2},
		},
	}

	for _, test := range tests {
		invSet := NewInventorySet(test.initialItems)
		actual := invSet.DeleteItem(test.deleteItem)
		if test.expected != actual {
			t.Errorf("Expected return value (%t), got (%t)\n", test.expected, actual)
		}
		if len(test.expectedItems) != invSet.Size() {
			t.Errorf("Expected num inventory items (%d), got (%d)\n", len(test.expectedItems), invSet.Size())
		}
	}
}

func TestInventorySetMerge(t *testing.T) {
	tests := []struct {
		set1   []*Inventory
		set2   []*Inventory
		merged []*Inventory
	}{
		{
			set1:   []*Inventory{},
			set2:   []*Inventory{},
			merged: []*Inventory{},
		},
		{
			set1:   []*Inventory{},
			set2:   []*Inventory{&inventory1},
			merged: []*Inventory{&inventory1},
		},
		{
			set1:   []*Inventory{&inventory1},
			set2:   []*Inventory{},
			merged: []*Inventory{&inventory1},
		},
		{
			set1:   []*Inventory{&inventory1, &inventory2},
			set2:   []*Inventory{&inventory1},
			merged: []*Inventory{&inventory1, &inventory2},
		},
		{
			set1:   []*Inventory{&inventory1, &inventory2},
			set2:   []*Inventory{&inventory1, &inventory2},
			merged: []*Inventory{&inventory1, &inventory2},
		},
		{
			set1:   []*Inventory{&inventory1, &inventory2},
			set2:   []*Inventory{&inventory3, &inventory4},
			merged: []*Inventory{&inventory1, &inventory2, &inventory3, &inventory4},
		},
	}

	for _, test := range tests {
		invSet1 := NewInventorySet(test.set1)
		invSet2 := NewInventorySet(test.set2)
		expected := NewInventorySet(test.merged)
		merged, _ := invSet1.Merge(invSet2)
		if expected.Size() != merged.Size() {
			t.Errorf("Expected merged inventory set size (%d), got (%d)\n", expected.Size(), merged.Size())
		}
	}
}

func TestInventorySetSubtract(t *testing.T) {
	tests := []struct {
		initialItems  []*Inventory
		subtractItems []*Inventory
		expected      []*Inventory
	}{
		{
			initialItems:  []*Inventory{},
			subtractItems: []*Inventory{},
			expected:      []*Inventory{},
		},
		{
			initialItems:  []*Inventory{},
			subtractItems: []*Inventory{&inventory1},
			expected:      []*Inventory{},
		},
		{
			initialItems:  []*Inventory{&inventory1},
			subtractItems: []*Inventory{},
			expected:      []*Inventory{&inventory1},
		},
		{
			initialItems:  []*Inventory{&inventory1, &inventory2},
			subtractItems: []*Inventory{&inventory1},
			expected:      []*Inventory{&inventory2},
		},
		{
			initialItems:  []*Inventory{&inventory1, &inventory2},
			subtractItems: []*Inventory{&inventory1, &inventory2},
			expected:      []*Inventory{},
		},
		{
			initialItems:  []*Inventory{&inventory1, &inventory2},
			subtractItems: []*Inventory{&inventory3, &inventory4},
			expected:      []*Inventory{&inventory1, &inventory2},
		},
	}

	for _, test := range tests {
		invInitialItems := NewInventorySet(test.initialItems)
		invSubtractItems := NewInventorySet(test.subtractItems)
		expected := NewInventorySet(test.expected)
		actual, _ := invInitialItems.Subtract(invSubtractItems)
		if expected.Size() != actual.Size() {
			t.Errorf("Expected subtracted inventory set size (%d), got (%d)\n", expected.Size(), actual.Size())
		}
	}
}

func TestInventorySetEquals(t *testing.T) {
	tests := []struct {
		set1    []*Inventory
		set2    []*Inventory
		isEqual bool
	}{
		{
			set1:    []*Inventory{},
			set2:    []*Inventory{&inventory1},
			isEqual: false,
		},
		{
			set1:    []*Inventory{&inventory1},
			set2:    []*Inventory{},
			isEqual: false,
		},
		{
			set1:    []*Inventory{&inventory1, &inventory2},
			set2:    []*Inventory{&inventory1},
			isEqual: false,
		},
		{
			set1:    []*Inventory{&inventory1, &inventory2},
			set2:    []*Inventory{&inventory3, &inventory4},
			isEqual: false,
		},
		// Empty sets are equal.
		{
			set1:    []*Inventory{},
			set2:    []*Inventory{},
			isEqual: true,
		},
		{
			set1:    []*Inventory{&inventory1},
			set2:    []*Inventory{&inventory1},
			isEqual: true,
		},
		// Ordering of the inventory items does not matter for equality.
		{
			set1:    []*Inventory{&inventory1, &inventory2},
			set2:    []*Inventory{&inventory2, &inventory1},
			isEqual: true,
		},
	}

	for _, test := range tests {
		invSet1 := NewInventorySet(test.set1)
		invSet2 := NewInventorySet(test.set2)
		if !invSet1.Equals(invSet2) && test.isEqual {
			t.Errorf("Expected equal inventory sets; got unequal (%s)/(%s)\n", invSet1, invSet2)
		}
		if invSet1.Equals(invSet2) && !test.isEqual {
			t.Errorf("Expected inequal inventory sets; got equal (%s)/(%s)\n", invSet1, invSet2)
		}
	}
}
