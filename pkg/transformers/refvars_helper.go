package transformers

import (
	"fmt"
)

// TODO(jeb): Investigate the usage of obj->json + jsonPatch
// deepMerge merges two objects together
func deepMerge(dstVal interface{}, srcVal interface{}) (interface{}, error) {
	switch srcVal.(type) {
	case nil:
		// preserving non nil dstVal
		return dstVal, nil
	case []interface{}:
		return castAndMergeSlice(dstVal, srcVal)
	case map[string]interface{}:
		return castAndMergeMap(dstVal, srcVal)
	default:
		// preserving non nil srcVal
		return srcVal, nil
	}
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
		// preserving non nil srcVal
		return srcVal, nil
	default:
		return dstVal, fmt.Errorf("Conflicting type. Unable to merge %T", dstValType)
	}
}

func deepMergeSlice(dstSlice []interface{}, srcSlice []interface{}) ([]interface{}, error) {
	if len(dstSlice) != len(srcSlice) {
		return dstSlice,
			fmt.Errorf("Conflicting arrays. Unable to merge array with conflicting length")
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
		// preserving non nil srcVal
		return srcVal, nil
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

// TODO(jeb): Investigate the usage of obj->json->obj conversion
// to replace those functions.
// deepCopy copies slice and map. Returns original value otherwise
func deepCopy(srcVal interface{}) interface{} {
	switch srcVal.(type) {
	case nil:
		return srcVal
	case []interface{}:
		srcSlice := srcVal.([]interface{})
		return deepCopySlice(srcSlice)
	case map[string]interface{}:
		srcMap := srcVal.(map[string]interface{})
		return deepCopyMap(srcMap)
	default:
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
