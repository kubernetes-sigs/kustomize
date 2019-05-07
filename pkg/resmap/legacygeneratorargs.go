package resmap

import (
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/types"
)

func GeneratorArgsFromKunstruct(k ifc.Kunstructured) (
	result types.GeneratorArgs, err error) {
	result.Name = k.GetName()
	// TODO: validate behavior values.
	result.Behavior, err = k.GetFieldValue("behavior")
	if err != nil {
		return
	}
	result.EnvSources, err = k.GetStringSlice("envFiles")
	if err != nil {
		return
	}
	result.FileSources, err = k.GetStringSlice("valueFiles")
	if err != nil {
		return
	}
	result.LiteralSources, err = k.GetStringSlice("literals")
	return
}
