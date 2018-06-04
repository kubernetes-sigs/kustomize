package transformers

import (
	"fmt"

	"github.com/kubernetes-sigs/kustomize/pkg/expansion"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type refvarTransformer struct {
	pathConfigs []PathConfig
	vars        map[string]string
}

func NewRefVarTransformer(vars map[string]string) (Transformer, error) {
	return &refvarTransformer{
		vars: vars,
		pathConfigs: []PathConfig{
			{
				GroupVersionKind: &schema.GroupVersionKind{Kind: "StatefulSet"},
				Path:             []string{"spec", "template", "spec", "initContainers", "command"},
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{Kind: "StatefulSet"},
				Path:             []string{"spec", "template", "spec", "containers", "command"},
			},
			{
				GroupVersionKind: &schema.GroupVersionKind{Kind: "Job"},
				Path:             []string{"spec", "template", "spec", "containers", "command"},
			},
		},
	}, nil
}

func (rv *refvarTransformer) Transform(resources resmap.ResMap) error {
	// Determine the final values of variables:
	//
	// 1.  Determine the final value of each variable:
	//     a.  If the variable's Value is set, expand the `$(var)` references to other
	//         variables in the .Value field; the sources of variables are the declared
	//         variables of the container and the service environment variables
	//     b.  If a source is defined for an environment variable, resolve the source
	// 2.  Create the container's environment in the order variables are declared
	// 3.  Add remaining service environment vars

	for GVKn := range resources {
		obj := resources[GVKn].Unstruct()
		objMap := obj.UnstructuredContent()
		for _, pc := range rv.pathConfigs {
			if !selectByGVK(GVKn.Gvk(), pc.GroupVersionKind) {
				continue
			}
			err := mutateField(objMap, pc.Path, false, func(in interface{}) (interface{}, error) {
				var (
					mappingFunc = expansion.MappingFuncFor(rv.vars)
				)
				switch vt := in.(type) {
				case []interface{}:
					var xs []string
					for _, a := range in.([]interface{}) {
						xs = append(xs, expansion.Expand(a.(string), mappingFunc))
					}
					return xs, nil
				case interface{}:
					s, ok := in.(string)
					if !ok {
						return nil, fmt.Errorf("%#v is expectd to be %T", in, s)
					}
					runtimeVal := expansion.Expand(s, mappingFunc)
					return runtimeVal, nil
				default:
					return "", fmt.Errorf("invalid type encountered %T", vt)
				}
				return "", fmt.Errorf("invalid type encountered")
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
