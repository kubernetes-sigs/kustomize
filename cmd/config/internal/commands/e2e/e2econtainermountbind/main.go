// Copyright 2023 The Kubernetes Authors.
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
	api := new(struct {
		Path string `json:"path" yaml:"template"`
	})
	// create the template
	readFn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		generated := []*yaml.RNode{}
		for range items {
			bytes, err := os.ReadFile(api.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			resources, err := yaml.Parse(string(bytes))
			if err != nil {
				return nil, fmt.Errorf("failed to parse: %w", err)
			}
			generated = append(generated, resources)
		}

		return generated, nil
	}
	p := framework.SimpleProcessor{Config: api, Filter: kio.FilterFunc(readFn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
