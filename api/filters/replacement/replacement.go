// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package replacement

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/internal/utils"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml_utils "sigs.k8s.io/kustomize/kyaml/utils"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Filter struct {
	Replacements []types.Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

// Filter replaces values of targets with values from sources
func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for i, r := range f.Replacements {
		if r.Source == nil || r.Targets == nil {
			return nil, fmt.Errorf("replacements must specify a source and at least one target")
		}
		value, err := getReplacement(nodes, &f.Replacements[i])
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

func getReplacement(nodes []*yaml.RNode, r *types.Replacement) (*yaml.RNode, error) {
	source, err := selectSourceNode(nodes, r.Source)
	if err != nil {
		return nil, err
	}

	if r.Source.FieldPath == "" {
		r.Source.FieldPath = types.DefaultReplacementFieldPath
	}
	fieldPath := kyaml_utils.SmarterPathSplitter(r.Source.FieldPath, ".")

	rn, err := source.Pipe(yaml.Lookup(fieldPath...))
	if err != nil {
		return nil, fmt.Errorf("error looking up replacement source: %w", err)
	}
	if rn.IsNilOrEmpty() {
		return nil, fmt.Errorf("fieldPath `%s` is missing for replacement source %s", r.Source.FieldPath, r.Source.ResId)
	}

	return getRefinedValue(r.Source.Options, rn)
}

// selectSourceNode finds the node that matches the selector, returning
// an error if multiple or none are found
func selectSourceNode(nodes []*yaml.RNode, selector *types.SourceSelector) (*yaml.RNode, error) {
	var matches []*yaml.RNode
	for _, n := range nodes {
		ids, err := utils.MakeResIds(n)
		if err != nil {
			return nil, fmt.Errorf("error getting node IDs: %w", err)
		}
		for _, id := range ids {
			if id.IsSelectedBy(selector.ResId) {
				if len(matches) > 0 {
					return nil, fmt.Errorf(
						"multiple matches for selector %s", selector)
				}
				matches = append(matches, n)
				break
			}
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("nothing selected by %s", selector)
	}
	return matches[0], nil
}

func getRefinedValue(options *types.FieldOptions, rn *yaml.RNode) (*yaml.RNode, error) {
	if options == nil || options.Delimiter == "" {
		return rn, nil
	}
	if rn.YNode().Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("delimiter option can only be used with scalar nodes")
	}
	value := strings.Split(yaml.GetValue(rn), options.Delimiter)
	if options.Index >= len(value) || options.Index < 0 {
		return nil, fmt.Errorf("options.index %d is out of bounds for value %s", options.Index, yaml.GetValue(rn))
	}
	n := rn.Copy()
	n.YNode().Value = value[options.Index]
	return n, nil
}

func applyReplacement(nodes []*yaml.RNode, value *yaml.RNode, targetSelectors []*types.TargetSelector) ([]*yaml.RNode, error) {
	for _, selector := range targetSelectors {
		if selector.Select == nil {
			return nil, errors.Errorf("target must specify resources to select")
		}
		if len(selector.FieldPaths) == 0 {
			selector.FieldPaths = []string{types.DefaultReplacementFieldPath}
		}
		for _, possibleTarget := range nodes {
			ids, err := utils.MakeResIds(possibleTarget)
			if err != nil {
				return nil, err
			}

			// filter targets by label and annotation selectors
			selectByAnnoAndLabel, err := selectByAnnoAndLabel(possibleTarget, selector)
			if err != nil {
				return nil, err
			}
			if !selectByAnnoAndLabel {
				continue
			}

			// filter targets by matching resource IDs
			for i, id := range ids {
				if id.IsSelectedBy(selector.Select.ResId) && !rejectId(selector.Reject, &ids[i]) {
					err := copyValueToTarget(possibleTarget, value, selector)
					if err != nil {
						return nil, err
					}
					break
				}
			}
		}
	}
	return nodes, nil
}

func selectByAnnoAndLabel(n *yaml.RNode, t *types.TargetSelector) (bool, error) {
	if matchesSelect, err := matchesAnnoAndLabelSelector(n, t.Select); !matchesSelect || err != nil {
		return false, err
	}
	for _, reject := range t.Reject {
		if reject.AnnotationSelector == "" && reject.LabelSelector == "" {
			continue
		}
		if m, err := matchesAnnoAndLabelSelector(n, reject); m || err != nil {
			return false, err
		}
	}
	return true, nil
}

func matchesAnnoAndLabelSelector(n *yaml.RNode, selector *types.Selector) (bool, error) {
	r := resource.Resource{RNode: *n}
	annoMatch, err := r.MatchesAnnotationSelector(selector.AnnotationSelector)
	if err != nil {
		return false, err
	}
	labelMatch, err := r.MatchesLabelSelector(selector.LabelSelector)
	if err != nil {
		return false, err
	}
	return annoMatch && labelMatch, nil
}

func rejectId(rejects []*types.Selector, id *resid.ResId) bool {
	for _, r := range rejects {
		if !r.ResId.IsEmpty() && id.IsSelectedBy(r.ResId) {
			return true
		}
	}
	return false
}

func copyValueToTarget(target *yaml.RNode, value *yaml.RNode, selector *types.TargetSelector) error {
	for _, fp := range selector.FieldPaths {
		var targetFields []*yaml.RNode
		var err error
		if selector.Options != nil && selector.Options.Create {
			targetFields, err = findOrCreateFields(target, fp, value.YNode().Kind)
		} else {
			targetFields, err = findMatchingFields(target, fp)
		}
		if err != nil {
			return err
		}
		for _, t := range targetFields {
			if err := setFieldValue(selector.Options, t, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// findMatchingFields returns all fields in target that match the given field path.
// If the field path does not already exist in the target, an error is returned.
func findMatchingFields(target *yaml.RNode, fp string) ([]*yaml.RNode, error) {
	fieldPath := kyaml_utils.SmarterPathSplitter(fp, ".")
	// may return multiple fields, always wrapped in a sequence node
	foundFieldSequence, lookupErr := target.Pipe(&yaml.PathMatcher{Path: fieldPath})
	if lookupErr != nil {
		return nil, fmt.Errorf("error finding field in replacement target: %w", lookupErr)
	}
	targetFields, err := foundFieldSequence.Elements()
	if err != nil {
		return nil, fmt.Errorf("error fetching elements in replacement target: %w", err)
	}
	if len(targetFields) == 0 {
		return nil, errors.Errorf("unable to find field %s in replacement target", fp)
	}
	return targetFields, nil
}

// findOrCreateFields updates the field(s) in the target that match the given field path.
// If the field path does not already exist in the target, it is created.
// Wildcard matching is supported, but creating intermediate list nodes is not.
func findOrCreateFields(target *yaml.RNode, fp string, createKind yaml.Kind) ([]*yaml.RNode, error) {
	split := kyaml_utils.SmarterPathSplitter(fp, ".")
	targetPaths := [][]string{split}
	var err error
	if strings.Contains(fp, "*") {
		targetPaths, err = expandPaths(target.Copy(), split)
		if err != nil {
			return nil, err
		}
	}
	var targetFields []*yaml.RNode
	for _, fieldPath := range targetPaths {
		createdField, createErr := target.Pipe(yaml.LookupCreate(createKind, fieldPath...))
		if createErr != nil {
			return nil, fmt.Errorf("error creating replacement node: %w", createErr)
		}
		if createdField == nil {
			return nil, errors.Errorf("unable to find or create field %s in replacement target", fp)
		}
		targetFields = append(targetFields, createdField)
	}
	return targetFields, nil
}

// expandPaths creates one or more concrete paths from an input fieldPath that may contain wildcards.
// Each wildcard must correspond to an existing SequenceNode in the target.
// This means that the full path up to the wildcard must exist in order for a result to be generated.
// Paths that do not contain wildcards, or paths to the right of the last wildcard, do not need to exist.
// This is because expandPaths is intended for use with functions that may create paths where possible
// (it is not possible to create the undefined number of sequence elements represented by a wildcard).
func expandPaths(target *yaml.RNode, fieldPath []string) ([][]string, error) {
	if len(fieldPath) == 0 {
		return [][]string{[]string{}}, nil
	}

	field, remainder := fieldPath[0], fieldPath[1:]
	var results [][]string
	var subTargets []*yaml.RNode
	var err error

	if field != "*" {
		element, err := yaml.PathGetter{Path: []string{field}}.Filter(target)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		newPath := []string{field}
		if element == nil {
			// slurp the rest of the path until we find another wildcard
			// no further wildcards can be expanded, since the child nodes won't already exist either
			// and cannot automatically be created
			for _, p := range remainder {
				if p == "*" {
					break
				}
				newPath = append(newPath, p)
			}
			return [][]string{newPath}, nil
		}
		rResults, rErr := expandPaths(element, remainder)
		if rErr != nil {
			return nil, rErr
		}
		for _, rResult := range rResults {
			results = append(results, append(newPath, rResult...))
		}
		return results, nil
	}

	subTargets, err = target.Elements()
	if err != nil {
		return nil, errors.WrapPrefixf(err, "wildcard must target a list")
	}

	for resultNo, element := range subTargets {
		rResults, rErr := expandPaths(element, remainder)
		if rErr != nil {
			return nil, rErr
		}
		newPath := []string{fmt.Sprintf("%d", resultNo)}
		for _, rResult := range rResults {
			results = append(results, append(newPath, rResult...))
		}
	}
	return results, nil
}

func setFieldValue(options *types.FieldOptions, targetField *yaml.RNode, value *yaml.RNode) error {
	value = value.Copy()
	if options != nil && options.Delimiter != "" {
		if targetField.YNode().Kind != yaml.ScalarNode {
			return fmt.Errorf("delimiter option can only be used with scalar nodes")
		}
		tv := strings.Split(targetField.YNode().Value, options.Delimiter)
		v := yaml.GetValue(value)
		// TODO: Add a way to remove an element
		switch {
		case options.Index < 0: // prefix
			tv = append([]string{v}, tv...)
		case options.Index >= len(tv): // suffix
			tv = append(tv, v)
		default: // replace an element
			tv[options.Index] = v
		}
		value.YNode().Value = strings.Join(tv, options.Delimiter)
	}

	if targetField.YNode().Kind == yaml.ScalarNode {
		// For scalar, only copy the value (leave any type intact to auto-convert int->string or string->int)
		targetField.YNode().Value = value.YNode().Value
	} else {
		targetField.SetYNode(value.YNode())
	}

	return nil
}
