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

package transformers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type mutateFunc func(interface{}) (interface{}, error)

//
func applyMutateFunc(m map[string]interface{}, pathToField []string, fns []mutateFunc) error {
	// Reached the leaves
	var err error
	for _, fn := range fns {
		m[pathToField[0]], err = fn(m[pathToField[0]])
		if err != nil {
			return err
		}
	}
	return nil
}

// Get the next node in the path.
// The method has to analyse if the field looks like yyy/_XX_/zzz.
// _XX_ is the pattern used to access XX element of array yyy
func getNextNodeInPath(m map[string]interface{}, pathToField []string, fns []mutateFunc) (interface{}, []string, error) {

	v := m[pathToField[0]]
	newPathToField := pathToField[1:]

	// Check if this can be an index field
	if len(newPathToField) == 0 {
		return v, newPathToField, nil
	}

	// Verify the syntax of the index field, i.e _XX_
	tmpIndex := strings.Replace(newPathToField[0], "_", "", -1)
	pathIndex, err := strconv.Atoi(tmpIndex)
	if err != nil {
		return v, newPathToField, nil
	}

	// Verify the value is an array
	sliceV, ok := v.([]interface{})
	if !ok {
		// index but no array
		return v, newPathToField, nil
	}

	// Check index seems to be out of bound
	if len(sliceV) <= pathIndex {
		return v, newPathToField, nil
	}

	v = sliceV[pathIndex]
	if len(newPathToField) == 1 {
		// We reached a leaf node. Let's mutate the field
		err := applyMutateFunc(m, pathToField, fns)
		return v, newPathToField, err
	}

	newPathToField = newPathToField[1:]
	return v, newPathToField, nil
}

func mutateField(
	m map[string]interface{},
	pathToField []string,
	createIfNotPresent bool,
	fns ...mutateFunc) error {
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

	// We reached an leaf node
	if len(pathToField) == 1 {
		return applyMutateFunc(m, pathToField, fns)
	}

	// Let's extract the next node in the path
	v, newPathToField, err := getNextNodeInPath(m, pathToField, fns)
	if (newPathToField == nil) || (err != nil) {
		return err
	}

	switch typedV := v.(type) {
	case nil:
		log.Printf(
			"nil value at `%s` ignored in mutation attempt",
			strings.Join(pathToField, "."))
		return nil
	case map[string]interface{}:
		return mutateField(typedV, newPathToField, createIfNotPresent, fns...)
	case []interface{}:
		for i := range typedV {
			item := typedV[i]
			typedItem, ok := item.(map[string]interface{})
			if !ok {
				log.Printf("mutating %s %s", pathToField, m)
				return fmt.Errorf("%#v is expected to be %T", item, typedItem)
			}
			err := mutateField(typedItem, newPathToField, createIfNotPresent, fns...)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		// This can happen if object does not allows have the same structure.
		// This can currently happen because the tree substitution did no occur first
		// log.Printf( "%#v value at `%s` ignored in mutation attempt",
		// 			    typedV, strings.Join(pathToField, "."))
		return nil
	}
}
