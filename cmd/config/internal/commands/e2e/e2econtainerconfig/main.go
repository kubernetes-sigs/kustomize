// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
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

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			// set the annotation on each resource item
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
	}
	p := framework.SimpleProcessor{Filter: kio.FilterFunc(fn), Config: functionConfig}
	cmd := command.Build(p, command.StandaloneDisabled, false)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
