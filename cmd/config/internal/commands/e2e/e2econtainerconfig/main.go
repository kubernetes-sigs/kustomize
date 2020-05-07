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
	Data Data `yaml:"data,omitempty"`
}

func main() {
	functionConfig := &Example{}

	cmd := framework.Command(functionConfig, func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			if err := items[i].PipeE(yaml.SetAnnotation("a-string-value",
				functionConfig.Data.StringValue)); err != nil {
				return nil, err
			}

			if err := items[i].PipeE(yaml.SetAnnotation("a-int-value",
				fmt.Sprintf("%v", functionConfig.Data.IntValue))); err != nil {
				return nil, err
			}

			if err := items[i].PipeE(yaml.SetAnnotation("a-bool-value",
				fmt.Sprintf("%v", functionConfig.Data.BoolValue))); err != nil {
				return nil, err
			}
		}
		return items, nil
	})

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
