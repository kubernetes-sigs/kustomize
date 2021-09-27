// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements a validator function run by `kustomize config run`
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/instrumenta/kubeval/kubeval"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	rw := &kio.ByteReadWriter{
		Reader:                os.Stdin,
		Writer:                os.Stdout,
		OmitReaderAnnotations: true,
		KeepReaderAnnotations: true,
	}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{rw}, // read the inputs into a slice
		Filters: []kio.Filter{kubevalFilter{rw: rw}},
		Outputs: []kio.Writer{rw}, // copy the inputs to the output
	}
	if err := p.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// kubevalFilter implements kio.Filter
type kubevalFilter struct {
	rw *kio.ByteReadWriter
}

// define the input API schema as a struct
type API struct {
	Spec struct {
		// Strict disallows additional properties not in schema if set.
		Strict bool `yaml:"strict"`

		// IgnoreMissingSchemas skips validation for resource
		// definitions without a schema.
		IgnoreMissingSchemas bool `yaml:"ignoreMissingSchemas"`

		// KubernetesVersion is the version of Kubernetes to validate
		// against (default "master").
		KubernetesVersion string `yaml:"kubernetesVersion"`

		// SchemaLocation is the base URL used to download schemas.
		SchemaLocation string `yaml:"schemaLocation"`
	} `yaml:"spec"`
}

// Filter checks each resource for validity, otherwise returning an error.
func (f kubevalFilter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	api := f.parseAPI()
	config := kubeval.NewDefaultConfig()
	config.Strict = api.Spec.Strict
	config.IgnoreMissingSchemas = api.Spec.IgnoreMissingSchemas
	config.KubernetesVersion = api.Spec.KubernetesVersion
	config.SchemaLocation = api.Spec.SchemaLocation

	// validate each Resource
	for _, r := range in {
		if err := validate(r.MustString(), config); err != nil {
			meta, merr := r.GetMeta()
			if merr != nil {
				return nil, merr
			}
			fmt.Fprintf(
				os.Stderr,
				"Resource invalid: (Kind: %s, Name: %s)\n",
				meta.Kind, meta.Name,
			)
			return nil, err
		}
	}
	return in, nil
}

func validate(r string, config *kubeval.Config) error {
	results, err := kubeval.Validate([]byte(r), config)
	if err != nil {
		return err
	}

	return checkResults(results)
}

func checkResults(results []kubeval.ValidationResult) error {
	if len(results) == 0 {
		return nil
	}

	errs := []string{}
	for _, r := range results {
		for _, e := range r.Errors {
			// Workaround a bug where the
			// "config.kubernetes.io/index" annotation value is a
			// number (invalid), and is still set on the Resource
			// regardless of OmitReaderAnnotations.
			// TODO: Remove this once the above issues are
			// resolved.
			if e.String() == "metadata.annotations: Invalid type. Expected: [string,null], given: integer" {
				continue
			}
			errs = append(errs, e.String())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

// parseAPI parses the functionConfig into an API struct.
func (f *kubevalFilter) parseAPI() API {
	// parse the input function config -- TODO: simplify this
	var api API
	if err := yaml.Unmarshal([]byte(f.rw.FunctionConfig.MustString()), &api); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	return api
}
