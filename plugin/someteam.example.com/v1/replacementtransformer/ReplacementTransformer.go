// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

var (
	pattern = regexp.MustCompile(`(\S+)\[(\S+)=(\S+)\]`)
)

// Find matching image declarations and replace
// the name, tag and/or digest.
type plugin struct {
	Replacements []types.Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Replacements = []types.Replacement{}
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	for _, r := range p.Replacements {
		if r.Source == nil {
			return fmt.Errorf("`from` must be specified in one replacement")
		}
		if r.Target == nil {
			return fmt.Errorf("`to` must be specified in one replacement")
		}
		count := 0
		if r.Source.ObjRef != nil {
			count += 1
		}
		if r.Source.Value != "" {
			count += 1
		}
		if count > 1 {
			return fmt.Errorf("only one of fieldref and value is allowed in one replacement")
		}
	}
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) (err error) {
	for _, r := range p.Replacements {
		var replacement interface{}
		if r.Source.ObjRef != nil {
			replacement, err = getReplacement(m, r.Source.ObjRef, r.Source.FieldRef)
			if err != nil {
				return err
			}
		}
		if r.Source.Value != "" {
			replacement = r.Source.Value
		}
		fmt.Printf("The replacement is %s\n", replacement)
		err = substitute(m, r.Target, replacement)
		if err != nil {
			return err
		}
	}
	return nil
}

func getReplacement(m resmap.ResMap, objRef *types.Target, fieldRef string) (interface{}, error) {
	s := types.Selector{
		Gvk:       objRef.Gvk,
		Name:      objRef.Name,
		Namespace: objRef.Namespace,
	}
	resources, err := m.Select(s)
	if err != nil {
		return "", err
	}
	if len(resources) > 1 {
		return "", fmt.Errorf("found more than one resources matching from %v", resources)
	}
	if len(resources) == 0 {
		return "", fmt.Errorf("failed to find one resource matching from %v", objRef)
	}
	if fieldRef == "" {
		fieldRef = ".metadata.name"
	}
	return resources[0].GetFieldValue(fieldRef)
}

func substitute(m resmap.ResMap, to *types.ReplTarget, replacement interface{}) error {
	resources, err := m.Select(*to.ObjRef)
	if err != nil {
		return err
	}
	for _, r := range resources {
		for _, p := range to.FieldRefs {
			pathSlice := strings.Split(p, ".")
			if err := updateField(r.Map(), pathSlice, replacement); err != nil {
				return err
			}
		}
	}
	return nil
}

func getFirstPathSegment(path string) (field string, key string, value string, array bool) {
	groups := pattern.FindStringSubmatch(path)
	if len(groups) != 4 {
		return path, "", "", false
	}
	return groups[1], groups[2], groups[3], groups[2] != ""
}

func updateField(m interface{}, pathToField []string, replacement interface{}) error {
	if len(pathToField) == 0 {
		return nil
	}

	switch typedM := m.(type) {
	case map[string]interface{}:
		return updateMapField(typedM, pathToField, replacement)
	case []interface{}:
		return updateSliceField(typedM, pathToField, replacement)
	default:
		return fmt.Errorf("%#v is not expected to be a primitive type", typedM)
	}
}

func updateMapField(m map[string]interface{}, pathToField []string, replacement interface{}) error {
	path, key, value, isArray := getFirstPathSegment(pathToField[0])

	v, found := m[path]
	if !found {
		m[path] = map[string]interface{}{}
		v = m[path]
	}

	if len(pathToField) == 1 {
		if !isArray {
			m[path] = replacement
			return nil
		}
		switch typedV := v.(type) {
		case nil:
			fmt.Printf("nil vlaue at `%s` ignored in mutation attempt", strings.Join(pathToField, "."))
		case []interface{}:
			for i := range typedV {
				item := typedV[i]
				typedItem, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("%#v is expected to be %T", item, typedItem)
				}
				if actualValue, ok := typedItem[key]; ok {
					if value == actualValue {
						typedItem[key] = value
					}
				}
			}
		default:
			return fmt.Errorf("%#v is not expected to be a primitive type", typedV)
		}
	}

	newPathToField := pathToField[1:]
	switch typedV := v.(type) {
	case nil:
		fmt.Printf(
			"nil value at `%s` ignored in mutation attempt",
			strings.Join(pathToField, "."))
		return nil
	case map[string]interface{}:
		return updateField(typedV, newPathToField, replacement)
	case []interface{}:
		if !isArray {
			return updateField(typedV, newPathToField, replacement)
		}
		for i := range typedV {
			item := typedV[i]
			typedItem, ok := item.(map[string]interface{})
			if !ok {
				return fmt.Errorf("%#v is expected to be %T", item, typedItem)
			}
			if actualValue, ok := typedItem[key]; ok {
				if value == actualValue {
					return updateField(typedItem, newPathToField, replacement)
				}
			}
		}
	default:
		return fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
	return nil
}

func updateSliceField(m []interface{}, pathToField []string, replacement interface{}) error {
	if len(pathToField) == 0 {
		return nil
	}
	index, err := strconv.Atoi(pathToField[0])
	if err != nil {
		return err
	}
	if len(m) > index && index >= 0 {
		if len(pathToField) == 1 {
			m[index] = replacement
			return nil
		} else {
			return updateField(m[index], pathToField[1:], replacement)
		}
	}
	return fmt.Errorf("index %v is out ouf bound", index)
}
