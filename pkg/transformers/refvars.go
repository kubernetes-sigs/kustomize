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
	"sigs.k8s.io/kustomize/v3/pkg/expansion"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/transformers/config"
)

type RefVarTransformer struct {
	varMap            map[string]interface{}
	replacementCounts map[string]int
	fieldSpecs        []config.FieldSpec
	mappingFunc       func(string) interface{}
}

const parentInline = "parent-inline"

// NewRefVarTransformer returns a new RefVarTransformer
// that replaces $(VAR) style variables with values.
// The fieldSpecs are the places to look for occurrences of $(VAR).
func NewRefVarTransformer(
	varMap map[string]interface{}, fs []config.FieldSpec) *RefVarTransformer {
	return &RefVarTransformer{
		varMap:     varMap,
		fieldSpecs: fs,
	}
}

// replaceStringField checks if the incoming field contains a string.
// If so, the field is processed to replace variables.
// If not, the field is returned as is
func (rv *RefVarTransformer) replaceStringField(a interface{}) interface{} {
	s, ok := a.(string)
	if !ok {
		// This field is not of string type.
		// It cannot contain a $(VAR)
		return a
	}

	// This field may contain a $(VAR)
	expandedValue := expansion.Expand(s, rv.mappingFunc)

	// Let's perform a deep copy if we didn't inline
	// a primitive type
	return deepCopy(expandedValue)
}

// inlineIntoParentNode allows to inline the complex tree of a variable into
// its parent node (as opposed to the current node).
// It is intended to be used as follow:
// ...
// construct1:
//   parent-field1:
//       parent-inline: $(var.pointing.to.a.shared.tree)
//       child-field1: value1
// ...
// construct2:
//   parent-field2:
//       parent-inline: $(var.pointing.to.a.shared.tree)
//       child-field2: value2
// ...
// Rationale: The simple inline of a variable map is quite often not
// enough to actually reduce copy/paste of yaml structs across documents.
// A user often needs to reuse an entire yaml tree, referred by the variable
// $(var.pointing.to.a.shared.tree) as a base across K8s constructs 1 and 2.
// He can then adjust the inlined content according to the needs of the
// current construct (child-field1 and child-field2)
func (rv *RefVarTransformer) inlineIntoParentNode(inMap map[string]interface{}) (interface{}, error) {
	s, _ := inMap[parentInline].(string)

	inlineValue := expansion.Expand(s, rv.mappingFunc)
	newMap, ok := inlineValue.(map[string]interface{})
	if !ok {
		return inMap, fmt.Errorf("parent-inline field must be expanded with a map[string]interface{}. Detected %s", inlineValue)
	}

	newMapCopy := deepCopyMap(newMap)
	mergedMap, err := deepMergeMap(newMapCopy, inMap)
	if err != nil {
		return inMap, fmt.Errorf("unable to merge current map %s into parent-inline map %s %v", inMap, newMap, err)
	}

	delete(mergedMap, parentInline)
	return mergedMap, nil
}

// replaceVars accepts as 'in' a string, or string array, which can have
// embedded instances of $VAR style variables, e.g. a container command string.
// The function returns the string with the variables expanded to their final
// values.
func (rv *RefVarTransformer) replaceVars(in interface{}) (interface{}, error) {
	switch in.(type) {
	case []interface{}:
		var xs []interface{}
		for _, a := range in.([]interface{}) {
			// Attempt to expand item by item
			xs = append(xs, rv.replaceStringField(a))
		}
		return xs, nil
	case map[string]interface{}:
		inMap := in.(map[string]interface{})

		// Deal with "parent-inline" special expansion
		if _, ok := inMap[parentInline]; ok {
			return rv.inlineIntoParentNode(inMap)
		}

		// Attempt to expand field by field
		xs := make(map[string]interface{}, len(inMap))
		for k, v := range inMap {
			xs[k] = rv.replaceStringField(v)
		}
		return xs, nil
	case string:
		// Attempt to expand this simple field
		return rv.replaceStringField(in), nil
	case nil:
		return nil, nil
	default:
		// This field cannot contain a $(VAR) since it is not of string type.
		return in, nil
	}
}

// UnusedVars returns slice of Var names that were unused
// after a Transform run.
func (rv *RefVarTransformer) UnusedVars() []string {
	var unused []string
	for k := range rv.varMap {
		_, ok := rv.replacementCounts[k]
		if !ok {
			unused = append(unused, k)
		}
	}
	return unused
}

// Transform replaces $(VAR) style variables with values.
func (rv *RefVarTransformer) Transform(m resmap.ResMap) error {
	rv.replacementCounts = make(map[string]int)
	rv.mappingFunc = expansion.MappingFuncFor(
		rv.replacementCounts, rv.varMap)

	// Then replace the variables. The first pass may inline
	// complex subtree, when the second can replace variables
	// reference inlined during the first pass
	for i := 0; i < 2; i++ {
		for _, res := range m.Resources() {
			for _, fieldSpec := range rv.fieldSpecs {
				if res.OrgId().IsSelected(&fieldSpec.Gvk) {
					if err := MutateField(
						res.Map(), fieldSpec.PathSlice(),
						false, rv.replaceVars); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
