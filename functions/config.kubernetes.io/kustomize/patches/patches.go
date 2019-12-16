// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package patches contains libraries for applying patches
package patches

import (
	"strings"

	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/util"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

var _ kio.Filter = &PatchFilter{}

// PatchFilter applies a set of patches
type PatchFilter struct {
	// Patches are the patches to apply
	Patches []Patch `yaml:"patches"`

	setterName string
}

// SetFieldSetter sets the name of the field metadata setter.
func (ns *PatchFilter) SetFieldSetter(n string) {
	ns.setterName = n
}

// Patch is a single patch to apply
type Patch struct {
	// Targets matches which Resources to apply the patch to
	Targets Targets `yaml:"targets,omitempty"`

	// Patch is the patch to apply
	Patch yaml.Node `yaml:"patch"`
}

// Targets configures how to match targets.
type Targets struct {
	// Kind is the kind of the Resources to match.  Optional.
	Kind string `yaml:"kind,omitempty"`

	// APIVersion is the apiVersion of the Resources to match.  Optional.
	APIVersion string `yaml:"apiVersion,omitempty"`

	// Name is the name of Resources to match.  Optional.
	Name string `yaml:"name,omitempty"`

	// Namespace is the namespace of Resources to match.  Optional.
	Namespace string `yaml:"namespace,omitempty"`

	// LabelSelector contains labels of Resources to match.  Optional.
	LabelSelector map[string]string `yaml:"labelSelector,omitempty"`

	// LabelPrefix matches Resources whose label value has the provided prefix.
	LabelPrefix map[string]string `yaml:"labelPrefix,omitempty"`

	// AnnotationSelector contains annotations of Resources to match.  Optional.
	AnnotationSelector map[string]string `yaml:"annotationSelector,omitempty"`

	// AnnotationPrefix matches Resources whose annotation value has the provided prefix.
	AnnotationPrefix map[string]string `yaml:"annotationPrefix,omitempty"`
}

func (pf *PatchFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range input {
		for j := range pf.Patches {
			match, err := pf.match(input[i], pf.Patches[j].Targets)
			if err != nil {
				return nil, err
			} else if !match {
				continue
			}

			// make a copy of the patch so when fields are copied to the destination,
			// each has a unique copy
			b, err := yaml.Marshal(&pf.Patches[j].Patch)
			if err != nil {
				return nil, err
			}
			var patch yaml.Node
			err = yaml.Unmarshal(b, &patch)
			if err != nil {
				return nil, err
			}

			v := yaml.NewRNode(patch.Content[0])
			// update the fieldmeta for the patched fields with the setter information
			if err := util.SetSetters(v, pf.setterName); err != nil {
				return nil, err
			}

			// merge the patch into the node
			input[i], err = merge2.Merge(v, input[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return input, nil
}

// match returns true if target matches node
func (pf *PatchFilter) match(node *yaml.RNode, target Targets) (bool, error) {
	meta, err := node.GetMeta()
	if err != nil {
		return false, err
	}
	if meta.Annotations == nil {
		meta.Annotations = map[string]string{}
	}
	if meta.Labels == nil {
		meta.Labels = map[string]string{}
	}

	if target.Name != "" && target.Name != meta.Name {
		return false, nil
	}
	if target.Kind != "" && target.Kind != meta.Kind {
		return false, nil
	}
	if target.APIVersion != "" && target.APIVersion != meta.APIVersion {
		return false, nil
	}
	if target.Namespace != "" && target.Namespace != meta.Namespace {
		return false, nil
	}

	for k, v := range target.AnnotationSelector {
		if meta.Annotations[k] != v {
			return false, nil
		}
	}
	for k, v := range target.AnnotationPrefix {
		if !strings.HasPrefix(meta.Annotations[k], v) {
			return false, nil
		}
	}
	for k, v := range target.LabelSelector {
		if meta.Labels[k] != v {
			return false, nil
		}
	}
	for k, v := range target.LabelPrefix {
		if !strings.HasPrefix(meta.Labels[k], v) {
			return false, nil
		}
	}
	return true, nil
}
