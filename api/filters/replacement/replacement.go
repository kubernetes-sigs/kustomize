// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package replacement

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Filter struct {
	Replacements []types.Replacement
}

// Filter replaces values of targets with values from sources
// TODO (#3492): Connect this to a replacement transformer plugin
func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, r := range f.Replacements {
		if r.Source == nil || r.Targets == nil {
			return nil, fmt.Errorf("replacements must specify a source and at least one target")
		}
		value, err := getReplacement(nodes, &r)
		if err != nil {
			return nil, err
		}
		nodes, err = applyReplacement(nodes, value, r.Targets)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func applyReplacement(nodes []*yaml.RNode, value *yaml.RNode, targets []*types.TargetSelector) ([]*yaml.RNode, error) {
	for _, t := range targets {
		if len(t.FieldPaths) == 0 {
			t.FieldPaths = []string{types.DefaultReplacementFieldPath}
		}
		for _, n := range nodes {
			// TODO (#3492): Don't include matches listed in the `reject` field
			if t.Select.KrmId.Match(getKrmId(n)) {
				err := applyToNode(n, value, t)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return nodes, nil
}

func applyToNode(node *yaml.RNode, value *yaml.RNode, target *types.TargetSelector) error {
	if target.Select == nil {
		return fmt.Errorf("target must specify resources to select")
	}
	for _, fp := range target.FieldPaths {
		fieldPath := strings.Split(fp, ".")
		// TODO (#3492): Add tests for map keys in the fieldPath (e.g. .spec.containers[name=nginx])
		t, err := node.Pipe(yaml.Lookup(fieldPath...))
		if err != nil {
			return err
		}
		if t != nil {
			// TODO (#3492): Use the field options to refine interpretation of the field
			t.SetYNode(value.YNode())
		}
	}
	return nil
}

func getReplacement(nodes []*yaml.RNode, r *types.Replacement) (*yaml.RNode, error) {
	source, err := selectSourceNode(nodes, r.Source)
	if err != nil {
		return nil, err
	}

	if r.Source.FieldPath == "" {
		r.Source.FieldPath = types.DefaultReplacementFieldPath
	}
	fieldPath := strings.Split(r.Source.FieldPath, ".")

	// TODO (#3492): Add tests for map keys in the fieldPath (e.g. .spec.containers[name=nginx])
	rn, err := source.Pipe(yaml.Lookup(fieldPath...))
	if err != nil {
		return nil, err
	}
	// TODO (#3492): Use the field options to refine interpretation of the field
	return rn, nil
}

// selectSourceNode finds the node that matches the selector, returning
// an error if multiple or none are found
func selectSourceNode(nodes []*yaml.RNode, selector *types.SourceSelector) (*yaml.RNode, error) {
	var matches []*yaml.RNode
	for _, n := range nodes {
		if selector.KrmId.Match(getKrmId(n)) {
			if len(matches) > 0 {
				return nil, fmt.Errorf("more than one match for source %v", selector)
			}
			matches = append(matches, n)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("found no matches for source %v", selector)
	}
	return matches[0], nil
}

func getKrmId(n *yaml.RNode) *types.KrmId {
	ns, _ := n.GetNamespace()
	apiVersion := yaml.GetValue(n.Field(yaml.APIVersionField).Value)
	group, version := resid.ParseGroupVersion(apiVersion)

	return &types.KrmId{
		Gvk:       resid.Gvk{Group: group, Version: version, Kind: n.GetKind()},
		Name:      n.GetName(),
		Namespace: ns,
	}
}
