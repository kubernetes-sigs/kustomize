// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Data contains the items
type Data struct {
	StringValue string `yaml:"stringValue,omitempty"`

	IntValue int `yaml:"intValue,omitempty"`

	BoolValue bool `yaml:"boolValue,omitempty"`
}

// Example defines the ResourceList.functionConfig schema.
type Example struct {
	// Data contains configuration data for the Example
	// Nest values under Data so that the function can accept a ConfigMap as its
	// functionConfig (`run` generates a ConfigMap for the functionConfig when run with --)
	// e.g. `config run DIR/ --image my-image -- a-string-value=foo` will create the input
	// with ResourceList.functionConfig.data.a-string-value=foo
	Data Data `yaml:"data,omitempty"`
}

func main() {
	functionConfig := &Example{}
	resourceList := &framework.ResourceList{FunctionConfig: functionConfig}

	cmd := framework.Command(resourceList, func() error {
		for i := range resourceList.Items {
			if err := resourceList.Items[i].PipeE(yaml.SetAnnotation("a-string-value",
				functionConfig.Data.StringValue)); err != nil {
				return err
			}

			if err := resourceList.Items[i].PipeE(yaml.SetAnnotation("a-int-value",
				fmt.Sprintf("%v", functionConfig.Data.IntValue))); err != nil {
				return err
			}

			if err := resourceList.Items[i].PipeE(yaml.SetAnnotation("a-bool-value",
				fmt.Sprintf("%v", functionConfig.Data.BoolValue))); err != nil {
				return err
			}
		}
		return nil
	})

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
