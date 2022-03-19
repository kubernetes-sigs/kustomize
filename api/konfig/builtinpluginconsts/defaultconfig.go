// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinpluginconsts

import (
	"bytes"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GetDefaultFieldSpecs returns default fieldSpecs.
func GetDefaultFieldSpecs() []byte {
	configData := [][]byte{
		[]byte(namePrefixFieldSpecs),
		[]byte(nameSuffixFieldSpecs),
		[]byte(commonLabelFieldSpecs),
		[]byte(commonAnnotationFieldSpecs),
		[]byte(namespaceFieldSpecs),
		[]byte(varReferenceFieldSpecs),
		[]byte(nameReferenceFieldSpecs),
		[]byte(imagesFieldSpecs),
		[]byte(replicasFieldSpecs),
	}
	return bytes.Join(configData, []byte("\n"))
}

// GetDefaultFieldSpecsAsMap returns default fieldSpecs
// as a string->string map.
func GetDefaultFieldSpecsAsMap() map[string]string {
	result := make(map[string]string)
	result["nameprefix"] = namePrefixFieldSpecs
	result["namesuffix"] = nameSuffixFieldSpecs
	result["commonlabels"] = commonLabelFieldSpecs
	result["commonannotations"] = commonAnnotationFieldSpecs
	result["namespace"] = namespaceFieldSpecs
	result["varreference"] = varReferenceFieldSpecs
	result["namereference"] = nameReferenceFieldSpecs
	result["images"] = imagesFieldSpecs
	result["replicas"] = replicasFieldSpecs
	return result
}

// GetFsSliceAsMap returns default fieldSpecs
// as a string->types.FsSlice map.
func GetFsSliceAsMap() (map[string]types.FsSlice, error) {
	result := make(map[string]types.FsSlice)
	for k, v := range GetDefaultFieldSpecsAsMap() {
		node, err := yaml.Parse(v)
		if err != nil {
			return nil, err
		}
		if err = node.VisitFields(func(mn *yaml.MapNode) error {
			var fsSlice types.FsSlice
			if err = mn.Value.VisitElements(func(elem *yaml.RNode) error {
				var fieldSpec types.FieldSpec
				if err := yaml.Unmarshal([]byte(elem.MustString()), &fieldSpec); err != nil {
					return err
				}
				fsSlice = append(fsSlice, fieldSpec)
				return nil
			}); err != nil {
				return err
			}
			result[k] = fsSlice
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return result, nil
}
