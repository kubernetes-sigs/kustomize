// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements an injection function for resource reservations and
// is run with `kustomize fn run -- DIR/`.
package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var (
	// cpuSizes is the mapping from tshirt-size to cpu reservation quantity
	cpuSizes = map[string]string{
		"small":  "200m",
		"medium": "4",
		"large":  "16",
	}

	// memorySizes is the mapping from tshirt-size to memory reservation quantity
	memorySizes = map[string]string{
		"small":  "50M",
		"medium": "1G",
		"large":  "32G",
	}
)

func main() {
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for _, item := range items {
			if err := size(item); err != nil {
				return nil, err
			}
		}
		return items, nil
	}
	p := framework.SimpleProcessor{Config: nil, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// size sets the cpu and memory reservations on all containers for Resources annotated
// with `tshirt-size: small|medium|large`
func size(r *yaml.RNode) error {
	// lookup the containers field
	containers, err := r.Pipe(yaml.Lookup("spec", "template", "spec", "containers"))
	if err != nil {
		s, _ := r.String()
		return fmt.Errorf("%v: %s", err, s)
	}
	if containers == nil {
		// doesn't have containers, skip the Resource
		return nil
	}

	// check for the tshirt-size annotations
	meta, err := r.GetMeta()
	if err != nil {
		return err
	}
	var memorySize, cpuSize string
	if size, found := meta.Annotations["tshirt-size"]; !found {
		// not a tshirt-sized Resource, ignore it
		return nil
	} else {
		// lookup the memory and cpu quantities based on the tshirt size
		memorySize = memorySizes[size]
		cpuSize = cpuSizes[size]
		if memorySize == "" || cpuSize == "" {
			return fmt.Errorf("unsupported tshirt-size: " + size)
		}
	}

	// visit each container and apply the cpu and memory reservations
	return containers.VisitElements(func(node *yaml.RNode) error {
		// set cpu
		err := node.PipeE(
			// lookup resources.requests.cpu, creating the field as a
			// ScalarNode if it doesn't exist
			yaml.LookupCreate(yaml.ScalarNode, "resources", "requests", "cpu"),
			// set the field value to the cpuSize
			yaml.Set(yaml.NewScalarRNode(cpuSize)))
		if err != nil {
			s, _ := r.String()
			return fmt.Errorf("%v: %s", err, s)
		}

		// set memory
		err = node.PipeE(
			// lookup resources.requests.memory, creating the field as a
			// ScalarNode if it doesn't exist
			yaml.LookupCreate(yaml.ScalarNode, "resources", "requests", "memory"),
			// set the field value to the memorySize
			yaml.Set(yaml.NewScalarRNode(memorySize)))
		if err != nil {
			s, _ := r.String()
			return fmt.Errorf("%v: %s", err, s)
		}

		return nil
	})
}
