package app

import (
	"fmt"

	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
)

func gvkn(rv types.Var) resource.ResId {
	return resource.NewResId(rv.ObjRef.GroupVersionKind(), rv.ObjRef.Name)
}

func getFieldAsString(m map[string]interface{}, pathToField []string) (string, error) {
	if len(pathToField) == 0 {
		return "", fmt.Errorf("Field not found")
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

	curr, rest := pathToField[0], pathToField[1]

	v := m[curr]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return getFieldAsString(typedV, []string{rest})
	default:
		return "", fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}
