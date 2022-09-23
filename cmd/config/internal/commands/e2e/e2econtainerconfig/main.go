// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"strconv"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Config defines the ResourceList.functionConfig schema.
type Config struct {
	// Data contains configuration data for the Example
	// Nest values under Data so that the function can accept a ConfigMap as its
	// functionConfig (`run` generates a ConfigMap for the functionConfig when run with --)
	// e.g. `config run DIR/ --image my-image -- a-string-value=foo` will create the input
	// with ResourceList.functionConfig.data.a-string-value=foo
	Data map[string]string `yaml:"data,omitempty"`
}

func main() {
	config := &Config{}

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		var newNodes []*yaml.RNode
		for i := range items {
			// set the annotation on each resource item
			if err := items[i].PipeE(yaml.SetAnnotation("a-string-value", config.Data["stringValue"])); err != nil {
				return nil, fmt.Errorf("%w", err)
			}
			intValue, err := strconv.Atoi(config.Data["intValue"])
			if config.Data["intValue"] == "" {
				// default value
				intValue = 0
			} else if err != nil {
				return nil, fmt.Errorf("intValue convert error: %w", err)

			}
			if err := items[i].PipeE(yaml.SetAnnotation("a-int-value", strconv.Itoa(intValue))); err != nil {
				return nil, fmt.Errorf("%w", err)
			}
			boolValue, err := strconv.ParseBool(config.Data["boolValue"])
			if config.Data["boolValue"] == "" {
				// default value
				boolValue = false
			} else if err != nil {
				return nil, fmt.Errorf("boolValue convert error: %w", err)

			}
			if err := items[i].PipeE(yaml.SetAnnotation("a-bool-value", strconv.FormatBool(boolValue))); err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			newNodes = append(newNodes, items[i])
		}
		return newNodes, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
