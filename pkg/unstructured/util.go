/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package unstructured

import (
	"fmt"
)

func getFieldValue(m map[string]interface{}, pathToField []string) (string, error) {
	if len(pathToField) == 0 {
		return "", fmt.Errorf("field not found")
	}
	if len(pathToField) == 1 {
		if v, found := m[pathToField[0]]; found {
			if s, ok := v.(string); ok {
				return s, nil
			}
			return "", fmt.Errorf("value at fieldpath is not of string type")
		}
		return "", fmt.Errorf("field at given fieldpath does not exist")
	}
	v := m[pathToField[0]]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return getFieldValue(typedV, pathToField[1:])
	default:
		return "", fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}

func getMapFieldValue(m map[string]interface{}, pathToField []string) (map[string]interface{}, error) {
	if len(pathToField) == 0 {
		return nil, fmt.Errorf("field not found")
	}
	if len(pathToField) == 1 {
		if v, found := m[pathToField[0]]; found {
			if s, ok := v.(map[string]interface{}); ok {
				return s, nil
			}
			return nil, fmt.Errorf("value at fieldpath is not of map type")
		}
		return nil, fmt.Errorf("field at given fieldpath does not exist")
	}
	v := m[pathToField[0]]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return getMapFieldValue(typedV, pathToField[1:])
	default:
		return nil, fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}

func mutateField(m map[string]interface{}, pathToField []string, createIfNotPresent bool, value string) error {
	if len(pathToField) == 0 {
		return nil
	}

	_, found := m[pathToField[0]]
	if !found {
		if !createIfNotPresent {
			return nil
		}
		m[pathToField[0]] = map[string]interface{}{}
	}

	if len(pathToField) == 1 {
		m[pathToField[0]] = value
		return nil
	}

	v := m[pathToField[0]]
	newPathToField := pathToField[1:]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return mutateField(typedV, newPathToField, createIfNotPresent, value)
	case []interface{}:
		for i := range typedV {
			item := typedV[i]
			typedItem, ok := item.(map[string]interface{})
			if !ok {
				return fmt.Errorf("%#v is expectd to be %T", item, typedItem)
			}
			err := mutateField(typedItem, newPathToField, createIfNotPresent, value)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}

func convertToStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if s, ok := v.(string); ok {
			result[k] = s
		} else {
			return nil
		}
	}
	return result
}

func convertToInterfaceMap(m map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

func setMapFieldValue(obj map[string]interface{}, pathToField []string, m map[string]interface{}) error {
	if len(pathToField) == 0 {
		return fmt.Errorf("field not found")
	}
	if len(pathToField) == 1 {
		obj[pathToField[0]] = m
	}
	v := obj[pathToField[0]]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return setMapFieldValue(typedV, pathToField[1:], m)
	default:
		return fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}
