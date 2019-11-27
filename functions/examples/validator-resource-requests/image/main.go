// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements a validator function run by `kustomize config run`
package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	rw := &kio.ByteReadWriter{Reader: os.Stdin, Writer: os.Stdout, KeepReaderAnnotations: true}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{rw},       // read the inputs into a slice
		Filters: []kio.Filter{filter{}}, // run the filter against the inputs
		Outputs: []kio.Writer{rw}}       // copy the inputs to the output
	if err := p.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// filter implements kio.Filter
type filter struct{}

// Filter checks each input and ensures that all containers have cpu and memory
// reservations set, otherwise it returns an error.
func (filter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	// validate each Resource
	for _, r := range in {
		if err := validate(r); err != nil {
			return nil, err
		}
	}
	return in, nil
}

func validate(r *yaml.RNode) error {
	meta, err := r.GetMeta()
	if err != nil {
		return err
	}

	// lookup the containers field in the Resource
	containers, err := r.Pipe(yaml.Lookup("spec", "template", "spec", "containers"))
	if err != nil {
		s, _ := r.String()
		return fmt.Errorf("%v: %s", err, s)
	}
	if containers == nil {
		// doesn't have containers, ignore it
		return nil
	}

	// visit each container in the list and validate
	return containers.VisitElements(func(node *yaml.RNode) error {

		// check cpu is non-nil
		f, err := node.Pipe(yaml.Lookup("resources", "requests", "cpu"))
		if err != nil {
			s, _ := r.String()
			return fmt.Errorf("%v: %s", err, s)
		}
		if f == nil {
			return fmt.Errorf(
				"cpu-requests missing for a container in %s %s (%s [%s])",
				meta.Kind, meta.Name,
				meta.Annotations[kioutil.PathAnnotation],
				meta.Annotations[kioutil.IndexAnnotation])
		}

		// check memory is non-nil
		f, err = node.Pipe(yaml.Lookup("resources", "requests", "memory"))
		if err != nil {
			s, _ := r.String()
			return fmt.Errorf("%v: %s", err, s)
		}
		if f == nil {
			return fmt.Errorf(
				"memory-requests missing for a container in %s in %s (%s [%s])",
				meta.Kind, meta.Name,
				meta.Annotations[kioutil.PathAnnotation],
				meta.Annotations[kioutil.IndexAnnotation])
		}

		// container is valid
		return nil
	})
}
