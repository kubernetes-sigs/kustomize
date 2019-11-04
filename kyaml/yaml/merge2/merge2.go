// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package merge contains libraries for merging fields from one RNode to another
// RNode
package merge2

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/walk"
)

const Help = `
Description:

  merge merges fields from a source to a destination, overriding the destination fields
  where they differ.

  ### Merge Rules

  Fields are recursively merged using the following rules:

  - scalars
    - if present only in the dest, it keeps its value
    - if present in the src and is non-null, take the src value -- if ` + "`null`" + `, clear it
` + "    - example src: `5`, dest: `3` => result: `5`" + `

  - non-associative lists -- lists without a merge key
    - if present only in the dest, it keeps its value
    - if present in the src and is non-null, take the src value -- if ` + "`null`" + `, clear it
` + "    - example src: `[1, 2, 3]`, dest: `[a, b, c]` => result: `[1, 2, 3]`" + `

  - map keys and fields -- paired by the map-key / field-name
    - if present only in the dest, it keeps its value
    - if present only in the src, it is added to the dest
    - if the field is present in both the src and dest, and the src value is 'null', the field is removed from the dest
    - if the field is present in both the src and dest, the value is recursively merged
` + "    - example src: `{'key1': 'value1', 'key2': 'value2'}`, dest: `{'key2': 'value0', 'key3': 'value3'}` => result: `{'key1': 'value1', 'key2': 'value2', 'key3': 'value3'}`" + `

  - associative list elements -- paired by the associative key
    - if present only in the dest, it keeps its value in the list
    - if present only in the src, it is added to the dest list
    - if the field is present in both the src and dest, the value is recursively merged

  ### Associative Keys

  Associative keys are used to identify "same" elements within 2 different lists, and merge them.
  The following fields are recognized as associative keys:

` + "[`mountPath`, `devicePath`, `ip`, `type`, `topologyKey`, `name`, `containerPort`]" + `

  Any lists where all of the elements contain associative keys will be merged as associative lists.

  ### Example

  > Source

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 3 # scalar
	  template:
	    spec:
	      containers:  # associative list -- (name)
	      - name: nginx
	        image: nginx:1.7
	        command: ['new_run.sh', 'arg1'] # non-associative list
	      - name: sidecar2
	        image: sidecar2:v1

  > Destination

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 1
	  template:
	    spec:
	      containers:
	      - name: nginx
	        image: nginx:1.6
	        command: ['old_run.sh', 'arg0']
	      - name: sidecar1
	        image: sidecar1:v1

  > Result

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 3 # scalar
	  template:
	    spec:
	      containers:  # associative list -- (name)
	      - name: nginx
	        image: nginx:1.7
	        command: ['new_run.sh', 'arg1'] # non-associative list
	      - name: sidecar1
	        image: sidecar1:v1
	      - name: sidecar2
	        image: sidecar2:v1
`

// Merge merges fields from src into dest.
func Merge(src, dest *yaml.RNode) (*yaml.RNode, error) {
	return walk.Walker{Sources: []*yaml.RNode{dest, src}, Visitor: Merger{}}.Walk()
}

// Merge parses the arguments, and merges fields from srcStr into destStr.
func MergeStrings(srcStr, destStr string) (string, error) {
	src, err := yaml.Parse(srcStr)
	if err != nil {
		return "", err
	}
	dest, err := yaml.Parse(destStr)
	if err != nil {
		return "", err
	}

	result, err := Merge(src, dest)
	if err != nil {
		return "", err
	}
	return result.String()
}

type Merger struct {
	// for forwards compatibility when new functions are added to the interface
}

var _ walk.Visitor = Merger{}

func (m Merger) VisitMap(nodes walk.Sources) (*yaml.RNode, error) {
	if err := m.SetComments(nodes); err != nil {
		return nil, err
	}
	if yaml.IsEmpty(nodes.Dest()) {
		// Add
		return nodes.Origin(), nil
	}
	if yaml.IsNull(nodes.Origin()) {
		// clear the value
		return walk.ClearNode, nil
	}
	// Recursively Merge dest
	return nodes.Dest(), nil
}

func (m Merger) VisitScalar(nodes walk.Sources) (*yaml.RNode, error) {
	if err := m.SetComments(nodes); err != nil {
		return nil, err
	}
	// Override value
	if nodes.Origin() != nil {
		return nodes.Origin(), nil
	}
	// Keep
	return nodes.Dest(), nil
}

func (m Merger) VisitList(nodes walk.Sources, kind walk.ListKind) (*yaml.RNode, error) {
	if err := m.SetComments(nodes); err != nil {
		return nil, err
	}
	if kind == walk.NonAssociateList {
		// Override value
		if nodes.Origin() != nil {
			return nodes.Origin(), nil
		}
		// Keep
		return nodes.Dest(), nil
	}

	// Add
	if yaml.IsEmpty(nodes.Dest()) {
		return nodes.Origin(), nil
	}
	// Clear
	if yaml.IsNull(nodes.Origin()) {
		return walk.ClearNode, nil
	}
	// Recursively Merge dest
	return nodes.Dest(), nil
}

// SetComments copies the dest comments to the source comments if they are present
// on the source.
func (m Merger) SetComments(sources walk.Sources) error {
	source := sources.Origin()
	dest := sources.Dest()
	if source != nil && source.YNode().FootComment != "" {
		dest.YNode().FootComment = source.YNode().FootComment
	}
	if source != nil && source.YNode().HeadComment != "" {
		dest.YNode().HeadComment = source.YNode().HeadComment
	}
	if source != nil && source.YNode().LineComment != "" {
		dest.YNode().LineComment = source.YNode().LineComment
	}
	return nil
}
