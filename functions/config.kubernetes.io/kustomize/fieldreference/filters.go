// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldreference

import (
	"strings"

	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/fieldspec"
	"sigs.k8s.io/kustomize/kyaml/inpututil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type ResetKustomizeNameFilter struct {
	From                    map[NameKey]string
	NameKustomizeName       string
	nameOriginalAnnotation  string
	nameReferenceAnnotation string
}

type SetKustomizeNameFilter struct {
	From                    map[NameKey]string
	NameKustomizeName       string
	nameOriginalAnnotation  string
	nameReferenceAnnotation string
	To                      map[NameKey]string

	// commonNamePrefix contains a prefix to be applied to all Resource metadata.name fields
	// and Resource references
	NamePrefix string `yaml:"commonNamePrefix,omitempty"`

	// commonNameSuffix contains a suffix to be applied to all Resource metadata.name fields
	// and Resource references
	NameSuffix string `yaml:"commonNameSuffix,omitempty"`

	// nameReferenceSubstitutions contains name substitution mappings that should be performed.
	PrefixMappings []NameMappings `yaml:"nameReferenceSubstitutions,omitempty"`
}

type NameMappings struct {
	// namePrefix is the name prefix to match for substitution.  May optionally contain
	// namePrefix and nameSuffix of the referenced Resource.
	NamePrefix string `yaml:"namePrefix,omitempty"`

	// newSuffix is replaces the current Resource reference suffix.
	NewSuffix string `yaml:"newSuffix,omitempty"`

	// kind restricts the substitutions to a specific kind of referenced Resource -- e.g.
	// Secret
	Kind string `yaml:"kind,omitempty"`
}

type KustomizeNameFilter struct {
	NameKustomizeName string

	// commonNamePrefix is a name prefix to apply to all Resource metadata.name fields
	// and all Resource references.
	NamePrefix string `yaml:"commonNamePrefix,omitempty"`

	// commonNameSuffix is a name suffix to apply to all Resource metadata.name fields
	// and all Resource references.
	NameSuffix string `yaml:"commonNameSuffix,omitempty"`
}

func (knf *KustomizeNameFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	reset := &ResetKustomizeNameFilter{
		From:              map[NameKey]string{},
		NameKustomizeName: knf.NameKustomizeName,
	}
	set := SetKustomizeNameFilter{
		From:              reset.From,
		NameKustomizeName: reset.NameKustomizeName,
		To:                map[NameKey]string{},
		NamePrefix:        knf.NamePrefix,
		NameSuffix:        knf.NameSuffix,
	}
	input, err := reset.Filter(input)
	if err != nil {
		return nil, err
	}
	return set.Filter(input)
}

var _ kio.Filter = &ResetKustomizeNameFilter{}
var _ kio.Filter = &SetKustomizeNameFilter{}

func (knf *SetKustomizeNameFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	// initialize the object
	knf.To = map[NameKey]string{}
	knf.From = map[NameKey]string{}
	knf.nameOriginalAnnotation = OriginalNameAnnotation + knf.NameKustomizeName
	knf.nameReferenceAnnotation = ReferenceNameAnnotation + knf.NameKustomizeName

	// kustomize names and references
	if err := knf.setNewNames(input); err != nil {
		return nil, err
	}
	if err := knf.setNewNameReferences(input); err != nil {
		return nil, err
	}

	return input, nil
}

var suffixLen = len("-kust0123456789")

// newName generates the new name given an old name
func (knf *SetKustomizeNameFilter) newName(name string) string {
	if knf.NamePrefix != "" {
		name = knf.NamePrefix + name
	}

	if knf.NameSuffix != "" {
		if GetNamePatternRegex().MatchString(name) {
			// set the name suffix before the generated part of the name
			// so we can recover the non-generated name easily
			name = name[:len(name)-suffixLen] + knf.NameSuffix + name[len(name)-suffixLen:]
		} else {
			name = name + knf.NameSuffix
		}
	}
	return name
}

func (knf *SetKustomizeNameFilter) setNewNameReferences(inputs []*yaml.RNode) error {
	for i := range getNameReferenceFieldSpecs().Items {
		spec := getNameReferenceFieldSpecs().Items[i]
		key := NameKey{Kind: spec.Kind}
		if err := inpututil.MapInputsE(inputs, func(r *yaml.RNode, meta yaml.ResourceMeta) error {
			return r.PipeE(&fieldspec.FieldSpecListFilter{
				FieldSpecList: fieldspec.FieldSpecList{Items: spec.FieldSpecs},
				SetValue: func(node *yaml.RNode) error {
					// set the value to the original name
					key.Name = node.YNode().Value
					if newName, found := knf.To[key]; found {
						// name was found in the map, set it
						return node.PipeE(yaml.Set(yaml.NewScalarRNode(newName)))
					}

					// explicit substitutions
					for _, k := range knf.PrefixMappings {
						if k.Kind != "" && spec.Kind != "" && k.Kind != spec.Kind {
							continue
						}
						newName := knf.NamePrefix + k.NamePrefix + knf.NameSuffix + k.NewSuffix
						// check if the prefix matches
						if strings.HasPrefix(node.YNode().Value, k.NamePrefix) {
							knf.To[NameKey{Kind: spec.Kind, Name: node.YNode().Value}] = newName
							return node.PipeE(yaml.Set(yaml.NewScalarRNode(newName)))
						}
						// check if the prefix matches -- add the namePrefix and nameSuffix to it
						if strings.HasPrefix(node.YNode().Value, knf.NamePrefix+k.NamePrefix+knf.NameSuffix) {
							knf.To[NameKey{Kind: spec.Kind, Name: node.YNode().Value}] = newName
							return node.PipeE(yaml.Set(yaml.NewScalarRNode(newName)))
						}
					}

					if GetNamePatternRegex().MatchString(node.YNode().Value) {
						matches := GetNamePatternRegex().FindStringSubmatch(node.YNode().Value)
						key.Name = matches[1]
						if newName, found := knf.To[key]; found {
							// name was found in the map, reset it
							return node.PipeE(yaml.Set(yaml.NewScalarRNode(newName)))
						}
					}
					return nil
				},
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

// setNewNames applies name mapping changes
// - set NamePrefix
// - set references to generated Resource names
func (knf *SetKustomizeNameFilter) setNewNames(inputs []*yaml.RNode) error {
	return inpututil.MapInputsE(inputs, func(r *yaml.RNode, meta yaml.ResourceMeta) error {
		// get the new name of the object
		newName := knf.newName(meta.Name)

		// set an index from the reference name to the new name
		knf.To[NameKey{Kind: meta.Kind, Name: meta.Name}] = newName

		// set an index with the suffix stripped for generated names
		if GetNamePatternRegex().MatchString(meta.Name) {
			// update the index for references to the reference name
			matches := GetNamePatternRegex().FindStringSubmatch(meta.Name)
			knf.To[NameKey{Kind: meta.Kind, Name: matches[1]}] = newName
		}

		return r.PipeE(
			yaml.Lookup("metadata", "name"),
			yaml.Set(yaml.NewScalarRNode(newName)))
	})
}

func (knf *ResetKustomizeNameFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	// initialize the object
	if knf.From == nil {
		knf.From = map[NameKey]string{}
	}
	knf.nameOriginalAnnotation = OriginalNameAnnotation + knf.NameKustomizeName
	knf.nameReferenceAnnotation = ReferenceNameAnnotation + knf.NameKustomizeName

	// reset names and references back to their pre-kustomized values
	if err := knf.resetNames(input); err != nil {
		return nil, err
	}
	if err := knf.resetNameReferences(input); err != nil {
		return nil, err
	}

	return input, nil
}

// resetNameReferences resets name references
func (knf *ResetKustomizeNameFilter) resetNameReferences(inputs []*yaml.RNode) error {
	// iterate over each known referenced Resource type
	for i := range getNameReferenceFieldSpecs().Items {
		spec := getNameReferenceFieldSpecs().Items[i]
		// iterate over each Resource
		if err := inpututil.MapInputsE(inputs, func(r *yaml.RNode, meta yaml.ResourceMeta) error {
			// iterate over each known reference location
			return r.PipeE(&fieldspec.FieldSpecListFilter{
				FieldSpecList: fieldspec.FieldSpecList{Items: spec.FieldSpecs},
				SetValue: func(node *yaml.RNode) error {
					// look for the reference in the index of found Resources
					key := NameKey{Kind: spec.Kind, Name: node.YNode().Value}
					if name, found := knf.From[key]; found {
						// found this Resource, reset the reference to the reset name
						return node.PipeE(yaml.Set(yaml.NewScalarRNode(name)))
					}

					// if the referenced Resource has a generated name, then the suffix
					// may have changed.  strip the suffix and lookup from the index.
					if GetNamePatternRegex().MatchString(node.YNode().Value) {
						matches := GetNamePatternRegex().FindStringSubmatch(node.YNode().Value)
						key.Name = matches[1]
						if originalName, found := knf.From[key]; found {
							// name was found in the map, reset it
							return node.PipeE(yaml.Set(yaml.NewScalarRNode(originalName)))
						}
					}
					return nil
				},
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

// resetNames resets each Resources name to its original name -- removing previous
// kustomizations to the Resource's name
func (knf *ResetKustomizeNameFilter) resetNames(inputs []*yaml.RNode) error {
	return inpututil.MapInputsE(inputs, func(r *yaml.RNode, meta yaml.ResourceMeta) error {
		// get the original name of the Resource
		originalName := meta.Annotations[knf.nameOriginalAnnotation]

		if originalName == "" {
			// name never modified, no need to reset -- instead add the annotation
			return r.PipeE(yaml.SetAnnotation(knf.nameOriginalAnnotation, meta.Name))
		}

		// get the reference name for the object, only specified for generated ConfigMaps
		// where the reference name (configmap-name) will be different from the original
		// name (configmap-kust123456)
		referenceName := meta.Annotations[knf.nameReferenceAnnotation]
		if referenceName == "" {
			referenceName = originalName
		}

		// set the index from the name, to the reference name so we can also
		// reset references to this Resource in other objects.
		knf.From[NameKey{Kind: meta.Kind, Name: meta.Name}] = referenceName

		// for generated objects, also set the index that strips the suffix
		// this is used to update the reference from one Resource to another
		// having a different suffix.
		if GetNamePatternRegex().MatchString(meta.Name) {
			// create the index against the generated name
			matches := GetNamePatternRegex().FindStringSubmatch(meta.Name)
			knf.From[NameKey{Kind: meta.Kind, Name: matches[1]}] = referenceName
		}

		// reset the name to the original value
		return r.PipeE(
			yaml.Lookup("metadata", "name"),
			yaml.Set(yaml.NewScalarRNode(originalName)))
	})
}
