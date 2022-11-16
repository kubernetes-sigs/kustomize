// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"strings"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/yaml"
)

// LegacyKust holds kustomization information, including legacy fields
type LegacyKust struct {
	fields        legacyFields
	legacyPatches overloads
}

// legacyFields holds all current kustomization fields and any legacy fields that are not overloaded,
// or, in other words, do not conflict with current ones
type legacyFields struct {
	types.Kustomization `json:",inline" yaml:",inline"`
	ImageTags           []types.Image `json:"imageTags,omitempty" yaml:"imageTags,omitempty"`
}

// overloads holds all legacy kustomization fields whose names conflict with current ones
type overloads struct {
	Patches []types.PatchStrategicMerge `json:"patches,omitempty" yaml:"patches,omitempty"`
}

// Unmarshal treats YAML data as a kustomization and fills lk with data.
// Unmarshal throws an error if data is an invalid kustomization.
func (lk *LegacyKust) Unmarshal(data []byte) error {
	var keys map[string]interface{}
	// check for duplicate fields
	err := yaml.UnmarshalStrict(data, &keys)
	if err != nil {
		return errors.WrapPrefixf(err, "invalid kustomization")
	}

	// remove legacy patches, which will not unmarshal correctly into a kustomization
	var legacyPatches []types.PatchStrategicMerge
	var isLegacy bool
	if patchesField, exists := keys["patches"]; exists {
		legacyPatches, isLegacy = getLegacyPatches(patchesField)
		if isLegacy {
			delete(keys, "patches")
		}
	}

	dataWithoutLegacy, err := yaml.Marshal(&keys)
	if err != nil {
		return errors.WrapPrefixf(err, "invalid kustomization without legacy patches")
	}
	var fields legacyFields
	// check for unknown fields
	err = yaml.UnmarshalStrict(dataWithoutLegacy, &fields)
	if err != nil {
		return errors.WrapPrefixf(err, "invalid kustomization fields")
	}
	errs := fields.Kustomization.EnforceFields()
	if len(errs) > 0 {
		return errors.Errorf("invalid kustomization: %s", strings.Join(errs, "\n"))
	}

	lk.fields = fields
	lk.legacyPatches.Patches = legacyPatches
	return nil
}

// Marshal returns lk serialized
func (lk *LegacyKust) Marshal() ([]byte, error) {
	fieldsContent, err := yaml.Marshal(&lk.fields)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "unable to serialize legacy kustomization fields")
	}
	if string(fieldsContent) == "{}\n" {
		fieldsContent = []byte{}
	}
	legacyPatchesContent, err := yaml.Marshal(&lk.legacyPatches)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "unable to serialize legacy patches")
	}
	if string(legacyPatchesContent) == "{}\n" {
		legacyPatchesContent = []byte{}
	}
	return append(fieldsContent, legacyPatchesContent...), nil
}
