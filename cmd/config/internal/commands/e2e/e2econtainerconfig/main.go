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

func main() {
	var config struct {
		// Data contains the items
		Data map[string]string `yaml:"data"`
	}

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			// set the annotation on each resource item
			if err := items[i].PipeE(yaml.SetAnnotation("value", config.Data["value"])); err != nil {
				return nil, fmt.Errorf("%w", err)
			}
		}
		return items, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
