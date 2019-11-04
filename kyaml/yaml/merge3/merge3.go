// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package merge contains libraries for merging fields from one RNode to another
// RNode
package merge3

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/walk"
)

const Help = `
Description:

  merge3 identifies changes between an original source + updated source and merges the result
  into a destination, overriding the destination fields where they have changed between
  original and updated.

  ### Resource MergeRules

  - Resources present in the original and deleted from the update are deleted.
  - Resources missing from the original and added in the update are added.
  - Resources present only in the dest are kept without changes.
  - Resources present in both the update and the dest are merged *original + update + dest => dest*.

  ### Field Merge Rules

  Fields are recursively merged using the following rules:

  - scalars
    - if present in either dest or updated and 'null', clear the value
    - if unchanged between original and updated, keep dest value
    - if changed between original and updated (added, deleted, changed), take the updated value

  - non-associative lists -- lists without a merge key
    - if present in either dest or updated and 'null', clear the value
    - if unchanged between original and updated, keep dest value
    - if changed between original and updated (added, deleted, changed), take the updated value

  - map keys and fields -- paired by the map-key / field-name
    - if present in either dest or updated and 'null', clear the value
    - if present only in the dest, it keeps its value
    - if not-present in the dest, add the delta between original-updated as a field
    - otherwise recursively merge the value between original, updated, dest

  - associative list elements -- paired by the associative key
    - if present only in the dest, it keeps its value
    - if not-present in the dest, add the delta between original-updated as a field
    - otherwise recursively merge the value between original, updated, dest

  ### Associative Keys

  Associative keys are used to identify "same" elements within 2 different lists, and merge them.
  The following fields are recognized as associative keys:

` + "[`mountPath`, `devicePath`, `ip`, `type`, `topologyKey`, `name`, `containerPort`]" + `

  Any lists where all of the elements contain associative keys will be merged as associative lists.
`

func Merge(dest, original, update *yaml.RNode) (*yaml.RNode, error) {
	// if update == nil && original != nil => declarative deletion

	return walk.Walker{Visitor: Visitor{},
		Sources: []*yaml.RNode{dest, original, update}}.Walk()
}

func MergeStrings(dest, original, update string) (string, error) {
	srcOriginal, err := yaml.Parse(original)
	if err != nil {
		return "", err
	}
	srcUpdated, err := yaml.Parse(update)
	if err != nil {
		return "", err
	}
	d, err := yaml.Parse(dest)
	if err != nil {
		return "", err
	}

	result, err := Merge(d, srcOriginal, srcUpdated)
	if err != nil {
		return "", err
	}
	return result.String()
}
