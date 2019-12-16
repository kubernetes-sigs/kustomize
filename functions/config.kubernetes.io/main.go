// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements kustomize-functions
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	cmd := &cobra.Command{
		Use:          "config-function",
		SilenceUsage: true, // don't print usage on an error
		RunE:         (&Dispatcher{}).RunE,
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// Dispatcher dispatches to the matching API
type Dispatcher struct {
	// IO hanldes reading / writing Resources
	IO *kio.ByteReadWriter
}

func (d *Dispatcher) RunE(_ *cobra.Command, _ []string) error {
	d.IO = &kio.ByteReadWriter{
		Reader:                os.Stdin,
		Writer:                os.Stdout,
		KeepReaderAnnotations: true,
	}

	return kio.Pipeline{
		Inputs: []kio.Reader{d.IO},
		Filters: []kio.Filter{
			d, // invoke the API
			&filters.MergeFilter{},
			&filters.FileSetter{FilenamePattern: filepath.Join("config", "%n.yaml")},
			&filters.FormatFilter{},
		},
		Outputs: []kio.Writer{d.IO},
	}.Execute()
}

// dispatchTable maps configFunction Kinds to implementations
var dispatchTable = map[string]map[string]func() kio.Filter{
	"config.kubernetes.io/v1alpha1": {
		kustomize.InlineKustomizationKind: kustomize.InlineFilter,
	},
}

func (d *Dispatcher) Filter(inputs []*yaml.RNode) ([]*yaml.RNode, error) {
	// parse the API meta to find which API is being invoked
	meta, err := d.IO.FunctionConfig.GetMeta()
	if err != nil {
		return nil, err
	}

	api, found := dispatchTable[meta.APIVersion]
	if !found {
		return nil, fmt.Errorf("unsupported apiVersion: %s", meta.APIVersion)
	}

	// find the implementation for this API
	fn := api[meta.Kind]
	if fn == nil {
		return nil, fmt.Errorf("unsupported API type: %s", meta.Kind)
	}

	// dispatch to the implementation
	fltr := fn()

	// initializes the object from the config
	if err := yaml.Unmarshal([]byte(d.IO.FunctionConfig.MustString()), fltr); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Fprintf(os.Stderr, "%s\n", d.IO.FunctionConfig.MustString())
		os.Exit(1)
	}
	return fltr.Filter(inputs)
}
