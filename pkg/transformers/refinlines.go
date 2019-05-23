package transformers

import (
	"fmt"
	"log"
	"sigs.k8s.io/kustomize/pkg/expansion"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

type RefInlineTransformer struct {
	inlineMap         map[string]interface{}
	replacementCounts map[string]int
	fieldSpecs        []config.FieldSpec
	inlineFunc        func(string) interface{}
}

// NewRefInlineTransformer returns a new RefInlineTransformer
// that replaces $(VAR) style variables with values.
// The fieldSpecs are the places to look for occurrences of $(VAR).
func NewRefInlineTransformer(
	inlineMap map[string]interface{}, fs []config.FieldSpec) *RefInlineTransformer {
	return &RefInlineTransformer{
		inlineMap:  inlineMap,
		fieldSpecs: fs,
	}
}

// replaceInlines accepts as 'in' a string, or string array, which can have
// embedded instances of $VAR style variables, e.g. a container command string.
// The function returns the string with the variables expanded to their final
// values.
func (rv *RefInlineTransformer) inline(in interface{}) (interface{}, error) {
	switch in.(type) {
	case map[string]interface{}:
		inMap := in.(map[string]interface{})
		if _, ok1 := inMap["parent-inline"]; ok1 {
			s, _ := inMap["parent-inline"].(string)

			inlineValue := expansion.Inline(s, rv.inlineFunc)
			newMap, ok2 := inlineValue.(map[string]interface{})
			if !ok2 {
				log.Printf("inlining issue with %s", inlineValue)
				return inMap, nil
			}

			newMapCopy := deepCopyMap(newMap)
			mergedMap, err := deepMergeMap(newMapCopy, inMap)
			if err != nil {
				log.Printf("deepMerging issue with %s %v", newMap, err)
				return inMap, nil
			}
			delete(mergedMap, "parent-inline")
			return mergedMap, nil
		}
		return inMap, nil

	case string:
		s, _ := in.(string)
		inlineValue := expansion.Inline(s, rv.inlineFunc)
		return deepCopy(inlineValue), nil
	default:
		// log.Printf("inlining issue with %T %s", vt, in)
		return in, nil
	}
}

// UnusedInlines returns slice of Inline names that were unused
// after a Transform run.
func (rv *RefInlineTransformer) UnusedInlines() []string {
	var unused []string
	for k := range rv.inlineMap {
		_, ok := rv.replacementCounts[k]
		if !ok {
			unused = append(unused, k)
		}
	}
	return unused
}

// Transform replaces $(VAR) style variables with values.
func (rv *RefInlineTransformer) Transform(m resmap.ResMap) error {
	rv.replacementCounts = make(map[string]int)
	rv.inlineFunc = expansion.InlineFuncFor(
		rv.replacementCounts, rv.inlineMap)
	for id, res := range m {
		for _, fieldSpec := range rv.fieldSpecs {
			if id.Gvk().IsSelected(&fieldSpec.Gvk) {
				if err := mutateField(
					res.Map(), fieldSpec.PathSlice(),
					false, rv.inline); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Those functions should be deleted once we figure out how to register
// the schema and leverage K8s strategicPatch
func deepMerge(dstVal interface{}, srcVal interface{}) (interface{}, error) {
	switch srcValType := srcVal.(type) {
	case nil:
		// preserving non nil type
		return dstVal, nil
	case []interface{}:
		return castAndMergeSlice(dstVal, srcVal)
	case map[string]interface{}:
		return castAndMergeMap(dstVal, srcVal)
	default:
		switch dstValType := dstVal.(type) {
		case nil:
			return srcVal, nil
		default:
			if dstValType != srcValType {
				// return dstVal,
				//	fmt.Errorf("Conflicting type. Unable to merge key %T %T", srcValType, dstValType)
				dstVal = srcVal
			} else {
				dstVal = srcVal
			}
		}
	}

	return dstVal, nil
}

func castAndMergeSlice(dstVal interface{}, srcVal interface{}) (interface{}, error) {
	switch dstValType := dstVal.(type) {
	case []interface{}:
		dstSlice := dstVal.([]interface{})
		srcSlice := srcVal.([]interface{})
		mergedSlice, err := deepMergeSlice(dstSlice, srcSlice)
		if err != nil {
			return dstVal, err
		}
		return mergedSlice, nil
	case nil:
		return srcVal, nil
	default:
		return dstVal, fmt.Errorf("Conflicting type. Unable to merge %T", dstValType)
	}
}

func deepMergeSlice(dstSlice []interface{}, srcSlice []interface{}) ([]interface{}, error) {
	if len(dstSlice) != len(srcSlice) {
		return dstSlice,
			fmt.Errorf("Conflicting arrays. Unable to merge array with conflicting lenght")
	}
	for id, dstVal := range srcSlice {
		mergedVal, err := deepMerge(dstVal, srcSlice[id])
		if err != nil {
			return dstSlice, err
		}
		dstSlice[id] = mergedVal
	}

	return dstSlice, nil
}

func castAndMergeMap(dstVal interface{}, srcVal interface{}) (interface{}, error) {
	switch dstValType := dstVal.(type) {
	case map[string]interface{}:
		dstMap := dstVal.(map[string]interface{})
		srcMap := srcVal.(map[string]interface{})
		mergedMap, err := deepMergeMap(dstMap, srcMap)
		if err != nil {
			return dstVal, err
		}
		return mergedMap, nil
	case nil:
		return dstVal, nil
	default:
		return dstVal,
			fmt.Errorf("Conflicting type. Unable to merge %T", dstValType)

	}
}

func deepMergeMap(dstMap map[string]interface{}, srcMap map[string]interface{}) (map[string]interface{}, error) {
	for key, srcVal := range srcMap {
		if dstVal, ok := dstMap[key]; ok {
			mergedVal, err := deepMerge(dstVal, srcVal)
			if err != nil {
				return dstMap, err
			}
			dstMap[key] = mergedVal
		} else {
			dstMap[key] = srcVal
		}
	}
	return dstMap, nil
}

// Those functions should be deleted once we figure out how to register
// the schema and leverage K8s strategicPatch
func deepCopy(srcVal interface{}) interface{} {
	switch srcVal.(type) {
	case nil:
		// preserving non nil type
		return srcVal
	case []interface{}:
		srcSlice := srcVal.([]interface{})
		return deepCopySlice(srcSlice)
	case map[string]interface{}:
		srcMap := srcVal.(map[string]interface{})
		return deepCopyMap(srcMap)
	default:
		//JEB: Probably boggus
		return srcVal
	}
}

func deepCopySlice(srcSlice []interface{}) []interface{} {
	dstSlice := make([]interface{}, len(srcSlice))
	for id, srcVal := range srcSlice {
		dstSlice[id] = deepCopy(srcVal)
	}
	return dstSlice
}

func deepCopyMap(srcMap map[string]interface{}) map[string]interface{} {
	dstMap := make(map[string]interface{})
	for key, srcVal := range srcMap {
		dstMap[key] = deepCopy(srcVal)
	}
	return dstMap
}
