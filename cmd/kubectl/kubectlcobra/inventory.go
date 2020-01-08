// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"fmt"
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
