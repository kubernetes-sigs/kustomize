// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

// CheckRequiredSettersSet iterates through all the setter definitions in openAPI
// schema and returns error if any of the setter has required field true and isSet false
func CheckRequiredSettersSet() error {
	for key := range openapi.Schema().Definitions {
		if strings.Contains(key, fieldmeta.SetterDefinitionPrefix) {
			val := openapi.Schema().Definitions[key]
			defExt, err := GetExtFromSchema(&val) // parse the extension out of the openAPI
			if err != nil {
				return errors.Wrap(err)
			}
			if defExt.Setter.Required && !defExt.Setter.IsSet {
				return errors.Errorf("setter %s is required but not set, "+
					"please set it to new value and try again", strings.TrimPrefix(key, fieldmeta.SetterDefinitionPrefix))
			}
		}
	}
	return nil
}
