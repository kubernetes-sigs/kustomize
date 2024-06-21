// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package replacement

import (
	"fmt"
	"strings"
	"regexp"

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
	if r.Source.FullText != ""{
		rn := yaml.NewScalarRNode(r.Source.FullText)
		return rn, nil
	}
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
		selectByAnnoAndLabel, err := rejectByAnnoAndLabel(n, selector.Reject)
		if err != nil {
			return nil, err
		}
		if !selectByAnnoAndLabel {
			continue
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
	value := []string{}
	if options.EndDelimiter == "" {
		value = strings.Split(yaml.GetValue(rn), options.Delimiter)
	} else {
		mapper := func(s string) string {
			s = strings.ReplaceAll(s, options.Delimiter, "")
			s = strings.ReplaceAll(s, options.EndDelimiter, "")
			return s
		}
		if options.Delimiter == "" {
			return nil, fmt.Errorf("delimiter needs to be set if enddelimiter is set")
		}
		re := regexp.MustCompile(regexp.QuoteMeta(options.Delimiter) + `(.*?)` + regexp.QuoteMeta(options.EndDelimiter))
		dv := re.FindAllString(yaml.GetValue(rn), -1)
		for _, s := range dv {
			value = append(value, mapper(s))
		}
	}
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
			selectByAnnoAndLabel, err := selectByAnnoAndLabel(possibleTarget, selector.Select, selector.Reject)
			if err != nil {
				return nil, err
			}
			if !selectByAnnoAndLabel {
				continue
			}

			// filter targets by matching resource IDs
			for _, id := range ids {
				if id.IsSelectedBy(selector.Select.ResId) && !containsRejectId(selector.Reject, ids) {
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

func selectByAnnoAndLabel(n *yaml.RNode, s *types.Selector, r []*types.Selector) (bool, error) {
	if matchesSelect, err := matchesAnnoAndLabelSelector(n, s); !matchesSelect || err != nil {
		return false, err
	}
	return rejectByAnnoAndLabel(n, r)
}

func rejectByAnnoAndLabel(n *yaml.RNode, r []*types.Selector) (bool, error) {
	for _, reject := range r {
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

func containsRejectId(rejects []*types.Selector, ids []resid.ResId) bool {
	for _, r := range rejects {
		if r.ResId.IsEmpty() {
			continue
		}
		for _, id := range ids {
			if id.IsSelectedBy(r.ResId) {
				return true
			}
		}
	}
	return false
}

func copyValueToTarget(target *yaml.RNode, value *yaml.RNode, selector *types.TargetSelector) error {
	for _, fp := range selector.FieldPaths {
		createKind := yaml.Kind(0) // do not create
		if selector.Options != nil && selector.Options.Create {
			createKind = value.YNode().Kind
		}
		targetFieldList, err := target.Pipe(&yaml.PathMatcher{
			Path:   kyaml_utils.SmarterPathSplitter(fp, "."),
			Create: createKind})
		if err != nil {
			return errors.WrapPrefixf(err, fieldRetrievalError(fp, createKind != 0))
		}
		targetFields, err := targetFieldList.Elements()
		if err != nil {
			return errors.WrapPrefixf(err, fieldRetrievalError(fp, createKind != 0))
		}
		if len(targetFields) == 0 {
			return errors.Errorf(fieldRetrievalError(fp, createKind != 0))
		}

		for _, t := range targetFields {
			if err := setFieldValue(selector.Options, t, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func fieldRetrievalError(fieldPath string, isCreate bool) string {
	if isCreate {
		return fmt.Sprintf("unable to find or create field %q in replacement target", fieldPath)
	}
	return fmt.Sprintf("unable to find field %q in replacement target", fieldPath)
}

func setFieldValue(options *types.FieldOptions, targetField *yaml.RNode, value *yaml.RNode) error {
	value = value.Copy()
	if options != nil && (options.Delimiter != "" || options.FullText != "") {
		if targetField.YNode().Kind != yaml.ScalarNode {
			return fmt.Errorf("delimiter option can only be used with scalar nodes")
		}
		v := yaml.GetValue(value)
		if options.FullText != "" {
			value.YNode().Value = getByRegex(options.FullText, targetField.YNode().Value, v, options.Index)
		} else if options.Delimiter != "" && options.EndDelimiter != "" {
			regex := regexp.QuoteMeta(options.Delimiter) + `(.*?)` + regexp.QuoteMeta(options.EndDelimiter)
			source := options.Delimiter + v + options.EndDelimiter
			value.YNode().Value = getByRegex(regex, targetField.YNode().Value, source, options.Index)
		} else {
			value.YNode().Value = getByDelimiter(options.Delimiter, targetField.YNode().Value, v, options.Index)
		}
	}

	if targetField.YNode().Kind == yaml.ScalarNode {
		// For scalar, only copy the value (leave any type intact to auto-convert int->string or string->int)
		targetField.YNode().Value = value.YNode().Value
	} else {
		targetField.SetYNode(value.YNode())
	}

	return nil
}

func getByDelimiter(delimiter string, target string, source string, index int) string {
	tv := strings.Split(target, delimiter)
	// TODO: Add a way to remove an element
	switch {
	case index < 0: // prefix
		tv = append([]string{source}, tv...)
	case index >= len(tv): // suffix
		tv = append(tv, source)
	default: // replace an element
		tv[index] = source
	}
	return strings.Join(tv, delimiter)
}

func getByRegex(regex string, target string, source string, index int) string {
	_, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("the regex: %s is not valid.", regex)
	}
	re := regexp.MustCompile(regex)
	counter := 0
	res := re.ReplaceAllStringFunc(target, func(str string) string {
		if counter != index && index >= 0 {
			return str
		}

		counter++
		return re.ReplaceAllString(str, source)
	})
	return res, nil
}

