// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements an injection function for resource reservations and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	rw := &kio.ByteReadWriter{Reader: os.Stdin, Writer: os.Stdout, KeepReaderAnnotations: true}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{rw},       // read the inputs into a slice
		Filters: []kio.Filter{filter{}}, // run the inject into the inputs
		Outputs: []kio.Writer{rw}}       // copy the inputs to the output
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

// filter implements kio.Filter
type filter struct{}

// Filter injects cpu and memory resource reservations into containers for
// Resources containing the `tshirt-size` annotation.
func (filter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	// inject the resource reservations into each Resource
	for _, r := range in {
		if err := inject(r); err != nil {
			return nil, err
		}
	}
	return in, nil
}

// cpuSizes is the mapping from tshirt-size to cpu reservation quantity
var cpuSizes = map[string]string{
	"small":  "200m",
	"medium": "4",
	"large":  "16",
}

// memorySizes is the mapping from tshirt-size to memory reservation quantity
var memorySizes = map[string]string{
	"small":  "50MiB",
	"medium": "1GiB",
	"large":  "32GiB",
}

// inject sets the cpu and memory reservations on all containers for Resources annotated
// with `tshirt-size: small|medium|large`
func inject(r *yaml.RNode) error {
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
