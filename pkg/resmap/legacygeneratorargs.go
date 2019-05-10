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
