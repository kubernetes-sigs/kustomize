// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const InlineKustomizationKind = "InlineKustomization"

// SetFieldSetter
func SetSetter(n *yaml.RNode, o string) error {
	if o == "" {
		// no-op
		return nil
	}
	fm := fieldmeta.FieldMeta{}
	if err := fm.Read(n); err != nil {
		return err
	}
	fm.OwnedBy = o
	return fm.Write(n)
}

func SetSetters(object *yaml.RNode, o string) error {
	return setSetters(object, o, true, false, "")
}

func setSetters(object *yaml.RNode, o string, root, meta bool, assc string) error {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		return setSetters(yaml.NewRNode(object.YNode().Content[0]), o, true, false, assc)
	case yaml.MappingNode:
		return object.VisitFields(func(node *yaml.MapNode) error {
			// special case unique scalars
			if node.Key.YNode().Value == assc {
				return nil
			}
			if root && node.Key.YNode().Value == "apiVersion" {
				return nil
			}
			if root && node.Key.YNode().Value == "kind" {
				return nil
			}
			if meta && node.Key.YNode().Value == "name" {
				return nil
			}
			if meta && node.Key.YNode().Value == "namespace" {
				return nil
			}
			if root && node.Key.YNode().Value == "metadata" {
				return setSetters(node.Value, o, false, true, "")
			}
			// no longer an associative key candidate
			return setSetters(node.Value, o, false, false, "")
		})
	case yaml.SequenceNode:
		// never set owner for associative keys -- the keys are shared across
		// all owners of the element
		key := object.GetAssociativeKey()
		return object.VisitElements(func(node *yaml.RNode) error {
			return setSetters(node, o, false, false, key)
		})
	case yaml.ScalarNode:
		return SetSetter(object, o)
	default:
		return nil
	}
}
