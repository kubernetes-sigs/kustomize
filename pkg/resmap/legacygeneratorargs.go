package resmap

import (
	"strings"

	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/types"
)

func GeneratorArgsFromKunstruct(k ifc.Kunstructured) (
	result types.GeneratorArgs, err error) {
	result.Name = k.GetName()
	// TODO: validate behavior values.
	result.Behavior, err = k.GetFieldValue("behavior")
	if !isAcceptableError(err) {
		return
	}
	result.EnvSources, err = k.GetStringSlice("envFiles")
	if !isAcceptableError(err) {
		return
	}
	result.FileSources, err = k.GetStringSlice("valueFiles")
	if !isAcceptableError(err) {
		return
	}
	result.LiteralSources, err = k.GetStringSlice("literals")
	if !isAcceptableError(err) {
		return
	}
	err = nil
	return
}

func isAcceptableError(err error) bool {
	return err == nil ||
		strings.HasPrefix(err.Error(), "no field named")
}
