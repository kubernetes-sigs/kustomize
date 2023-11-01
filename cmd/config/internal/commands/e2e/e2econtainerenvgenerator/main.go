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
	api := new(struct {
		Template string `json:"template" yaml:"template"`
	})
	// create the template
	templateFn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		generated := []*yaml.RNode{}
		for range items {
			templateResult, err := yaml.Parse(fmt.Sprintf(api.Template, os.Getenv("TESTTEMPLATE")))
			if err != nil {
				return nil, fmt.Errorf("failed to parse: %w", err)
			}
			generated = append(generated, templateResult)
		}
		return generated, nil
	}
	p := framework.SimpleProcessor{Config: api, Filter: kio.FilterFunc(templateFn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
