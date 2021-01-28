// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// SubstUtil holds the information about setters and substitutions
type SubstUtil struct {
	// SubstInfos information about all the substitutions which can be
	// created in the package
	SubstInfos []SubstInfo

	// SetterInfos is the information regarding the setters already present
	// in the package
	SetterInfos []SetterInfo
}

type SubstInfo struct {
	// FieldValue value of the field to create substitution for
	FieldValue string

	// Pattern setter pattern corresponding to FieldValue
	Pattern string
}

type SetterInfo struct {
	// SetterName name of setter present in package
	SetterName string

	// SetterValue value of setter SetterName present in package
	SetterValue string
}

// Filter implements yaml.Filter
func (s *SubstUtil) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	if len(s.SetterInfos) == 0 {
		return object, nil
	}
	return object, accept(s, object, nil)
}

func (s *SubstUtil) visitSequence(_ *yaml.RNode, _ string, _ *openapi.ResourceSchema) error {
	// no-op
	return nil
}

// visitMapping implements visitor
func (s *SubstUtil) visitMapping(object *yaml.RNode, p string, _ *openapi.ResourceSchema) error {
	// no-op
	return nil
}

// visitScalar implements visitor
// visitScalar visits all the scalar fields and populates SubstInfos with substitutions
// which can be created in the package using available setters
func (s *SubstUtil) visitScalar(object *yaml.RNode, p string, _, _ *openapi.ResourceSchema) error {
	if object.YNode().LineComment != "" {
		return nil
	}

	fieldValue := object.YNode().Value
	for _, substInfo := range s.SubstInfos {
		if fieldValue == substInfo.FieldValue {
			// fieldValue is already visited, so skip processing it again
			return nil
		}
	}
	pattern := patternFromValue(fieldValue, s.SetterInfos)
	if pattern != "" {
		s.SubstInfos = append(s.SubstInfos, SubstInfo{
			FieldValue: fieldValue,
			Pattern:    pattern,
		})
	}
	return nil
}

// patternFromValue takes the fieldValue of a scalar node and returns the
// corresponding pattern to create substitution which can be created using
// available setters in the package
func patternFromValue(fieldValue string, setterInfos []SetterInfo) string {
	for _, setterInfo := range setterInfos {
		if fieldValue == setterInfo.SetterValue {
			continue
		}
		fieldValue = strings.ReplaceAll(fieldValue, setterInfo.SetterValue, fmt.Sprintf("${%s}", setterInfo.SetterName))
	}
	if !strings.Contains(fieldValue, "${") {
		return ""
	}
	return fieldValue
}
