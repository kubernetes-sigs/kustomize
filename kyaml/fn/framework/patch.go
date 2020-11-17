// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

// Applier applies some modification to a ResourceList
type Applier interface {
	Apply(rl *ResourceList) error
}

var _ Applier = PatchTemplate{}

// PatchTemplate applies a patch to a collection of Resources
type PatchTemplate struct {
	// Template is a template to render into one or more patches.
	Template *template.Template

	// Selector targets the rendered patch to specific resources.
	Selector *Selector
}

// Apply applies the patch to all matching resources in the list.  The rl.FunctionConfig
// is provided to the template as input.
func (p PatchTemplate) Apply(rl *ResourceList) error {
	if p.Selector == nil {
		// programming error -- user shouldn't see this
		return errors.Errorf("must specify PatchTemplate.Selector")
	}

	matches, err := p.Selector.GetMatches(rl)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return nil
	}
	return p.apply(rl, p.Template, matches)
}

func (p *PatchTemplate) apply(rl *ResourceList, t *template.Template, matches []*yaml.RNode) error {
	// render the patches
	var b bytes.Buffer
	if err := t.Execute(&b, rl.FunctionConfig); err != nil {
		return errors.WrapPrefixf(err, "failed to render patch template %v", t.DefinedTemplates())
	}

	// parse the patches into RNodes
	var nodes []*yaml.RNode
	for _, s := range strings.Split(b.String(), "\n---\n") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		newNodes, err := (&kio.ByteReader{Reader: bytes.NewBufferString(s)}).Read()
		if err != nil {
			// create the debug string
			lines := strings.Split(s, "\n")
			for j := range lines {
				lines[j] = fmt.Sprintf("%03d %s", j+1, lines[j])
			}
			s = strings.Join(lines, "\n")
			return errors.WrapPrefixf(err, "failed to parse rendered patch template into a resource:\n%s\n", s)
		}
		nodes = append(nodes, newNodes...)
	}

	// apply the patches to the matching resources
	var err error
	for j := range matches {
		for i := range nodes {
			matches[j], err = merge2.Merge(nodes[i], p.Selector.matches[j], yaml.MergeOptions{})
			if err != nil {
				return errors.WrapPrefixf(err, "failed to merge templated patch")
			}
		}
	}
	return nil
}

// Selector matches resources.  A resource matches if and only if ALL of the Selector fields
// match the resource.  An empty Selector matches all resources.
type Selector struct {
	// Names is a list of metadata.names to match.  If empty match all names.
	// e.g. Names: ["foo", "bar"] matches if `metadata.name` is either "foo" or "bar".
	Names []string `json:"names" yaml:"names"`

	namesSet sets.String

	// Namespaces is a list of metadata.namespaces to match.  If empty match all namespaces.
	// e.g. Namespaces: ["foo", "bar"] matches if `metadata.namespace` is either "foo" or "bar".
	Namespaces []string `json:"namespaces" yaml:"namespaces"`

	namespaceSet sets.String

	// Kinds is a list of kinds to match.  If empty match all kinds.
	// e.g. Kinds: ["foo", "bar"] matches if `kind` is either "foo" or "bar".
	Kinds []string `json:"kinds" yaml:"kinds"`

	kindsSet sets.String

	// APIVersions is a list of apiVersions to match.  If empty apply match all apiVersions.
	// e.g. APIVersions: ["foo/v1", "bar/v1"] matches if `apiVersion` is either "foo/v1" or "bar/v1".
	APIVersions []string `json:"apiVersions" yaml:"apiVersions"`

	apiVersionsSet sets.String

	// Labels is a collection of labels to match.  All labels must match exactly.
	// e.g. Labels: {"foo": "bar", "baz": "buz"] matches if BOTH "foo" and "baz" labels match.
	Labels map[string]string `json:"labels" yaml:"labels"`

	// Annotations is a collection of annotations to match.  All annotations must match exactly.
	// e.g. Annotations: {"foo": "bar", "baz": "buz"] matches if BOTH "foo" and "baz" annotations match.
	Annotations map[string]string `json:"annotations" yaml:"annotations"`

	// Filter is an arbitrary filter function to match a resource.
	// Selector matches if the function returns true.
	Filter func(*yaml.RNode) bool

	// matches contains a list of matching reosurces.
	matches []*yaml.RNode

	// TemplatizeValues if set to true will parse the selector values as templates
	// and execute them with the functionConfig
	TemplatizeValues bool
}

// GetMatches returns them matching resources from rl
func (s *Selector) GetMatches(rl *ResourceList) ([]*yaml.RNode, error) {
	if err := s.init(rl); err != nil {
		return nil, err
	}
	return s.matches, nil
}

// templatize templatizes the value
func (s *Selector) templatize(value string, api interface{}) (string, error) {
	t, err := template.New("kinds").Parse(value)
	if err != nil {
		return "", errors.Wrap(err)
	}
	var b bytes.Buffer
	err = t.Execute(&b, api)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return b.String(), nil
}

func (s *Selector) templatizeSlice(values []string, api interface{}) error {
	var err error
	for i := range values {
		values[i], err = s.templatize(values[i], api)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector) templatizeMap(values map[string]string, api interface{}) error {
	var err error
	for k := range values {
		values[k], err = s.templatize(values[k], api)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector) init(rl *ResourceList) error {
	if s.TemplatizeValues {
		// templatize the selector values from the input configuration
		if err := s.templatizeSlice(s.Kinds, rl.FunctionConfig); err != nil {
			return err
		}
		if err := s.templatizeSlice(s.APIVersions, rl.FunctionConfig); err != nil {
			return err
		}
		if err := s.templatizeSlice(s.Names, rl.FunctionConfig); err != nil {
			return err
		}
		if err := s.templatizeSlice(s.Namespaces, rl.FunctionConfig); err != nil {
			return err
		}
		if err := s.templatizeMap(s.Labels, rl.FunctionConfig); err != nil {
			return err
		}
		if err := s.templatizeMap(s.Annotations, rl.FunctionConfig); err != nil {
			return err
		}
	}

	// index the selectors
	s.matches = nil
	s.kindsSet = sets.String{}
	s.kindsSet.Insert(s.Kinds...)
	s.apiVersionsSet = sets.String{}
	s.apiVersionsSet.Insert(s.APIVersions...)
	s.namesSet = sets.String{}
	s.namesSet.Insert(s.Names...)
	s.namespaceSet = sets.String{}
	s.namespaceSet.Insert(s.Namespaces...)

	// check each resource that matches the patch selector
	for i := range rl.Items {
		if match, err := s.isMatch(rl.Items[i]); err != nil {
			return err
		} else if !match {
			continue
		}
		s.matches = append(s.matches, rl.Items[i])
	}
	return nil
}

// isMatch returns true if r matches the patch selector
func (s *Selector) isMatch(r *yaml.RNode) (bool, error) {
	m, err := r.GetMeta()
	if err != nil {
		return false, errors.Wrap(err)
	}
	if s.kindsSet.Len() > 0 && !s.kindsSet.Has(m.Kind) {
		return false, nil
	}
	if s.apiVersionsSet.Len() > 0 && !s.apiVersionsSet.Has(m.APIVersion) {
		return false, nil
	}
	if s.namesSet.Len() > 0 && !s.namesSet.Has(m.Name) {
		return false, nil
	}
	if s.namespaceSet.Len() > 0 && !s.namespaceSet.Has(m.Namespace) {
		return false, nil
	}
	for k := range s.Labels {
		if m.Labels == nil || m.Labels[k] != s.Labels[k] {
			return false, nil
		}
	}
	for k := range s.Annotations {
		if m.Annotations == nil || m.Annotations[k] != s.Annotations[k] {
			return false, nil
		}
	}
	if s.Filter != nil {
		if match := s.Filter(r); !match {
			return false, nil
		}
	}
	return true, nil
}
