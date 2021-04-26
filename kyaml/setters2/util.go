// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"strings"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
)

// CheckRequiredSettersSet iterates through all the setter definitions in openAPI
// schema and returns error if any of the setter has required field true and isSet false
func CheckRequiredSettersSet(settersSchema *spec.Schema) error {
	for key := range settersSchema.Definitions {
		if strings.Contains(key, fieldmeta.SetterDefinitionPrefix) {
			val := settersSchema.Definitions[key]
			defExt, err := GetExtFromSchema(&val) // parse the extension out of the openAPI
			if err != nil {
				return errors.Wrap(err)
			}
			if defExt.Setter != nil && defExt.Setter.Required && !defExt.Setter.IsSet {
				return errors.Errorf("setter %s is required but not set, "+
					"please set it to new value and try again", strings.TrimPrefix(key, fieldmeta.SetterDefinitionPrefix))
			}
		}
	}
	return nil
}
