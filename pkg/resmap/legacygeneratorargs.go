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
	if !IsAcceptableError(err) {
		return
	}
	result.EnvSources, err = k.GetStringSlice("envFiles")
	if !IsAcceptableError(err) {
		return
	}
	result.FileSources, err = k.GetStringSlice("valueFiles")
	if !IsAcceptableError(err) {
		return
	}
	result.LiteralSources, err = k.GetStringSlice("literals")
	if !IsAcceptableError(err) {
		return
	}
	err = nil
	return
}

func IsAcceptableError(err error) bool {
	return err == nil ||
		strings.HasPrefix(err.Error(), "no field named")
}
