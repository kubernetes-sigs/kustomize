// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package runfn

import (
	"io"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// RunFns runs the set of configuration functions in a local directory against
// the Resources in that directory
type RunFns struct {
	// Path is the path to the directory containing functions
	Path string

	// FunctionPaths Paths allows functions to be specified outside the configuration
	// directory
	FunctionPaths []string

	// Output can be set to write the result to Output rather than back to the directory
	Output io.Writer

	// containerFilterProvider may be override by tests to fake invoking containers
	containerFilterProvider func(string, string, *yaml.RNode) kio.Filter
}

// Execute runs the command
func (r RunFns) Execute() error {
	// make the path absolute so it works on mac
	var err error
	r.Path, err = filepath.Abs(r.Path)
	if err != nil {
		return errors.Wrap(err)
	}

	// default the containerFilterProvider if it hasn't been override.  Split out for testing.
	(&r).init()

	// identify the configuration functions in the directory
	buff := &kio.PackageBuffer{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{kio.LocalPackageReader{PackagePath: r.Path}},
		Filters: []kio.Filter{&filters.IsReconcilerFilter{}},
		Outputs: []kio.Writer{buff},
	}.Execute()
	if err != nil {
		return err
	}

	// accept a
	for i := range r.FunctionPaths {
		err := kio.Pipeline{
			Inputs:  []kio.Reader{kio.LocalPackageReader{PackagePath: r.FunctionPaths[i]}},
			Outputs: []kio.Writer{buff},
		}.Execute()
		if err != nil {
			return err
		}
	}

	// reconcile each local API
	var fltrs []kio.Filter
	for i := range buff.Nodes {
		api := buff.Nodes[i]
		img, path := filters.GetContainerName(api)
		fltrs = append(fltrs, r.containerFilterProvider(img, path, api))
	}

	pkgIO := &kio.LocalPackageReadWriter{PackagePath: r.Path}
	inputs := []kio.Reader{pkgIO}
	var outputs []kio.Writer
	if r.Output == nil {
		// write back to the package
		outputs = append(outputs, pkgIO)
	} else {
		// write to the output instead of the directory
		outputs = append(outputs, kio.ByteWriter{Writer: r.Output})
	}
	return kio.Pipeline{Inputs: inputs, Filters: fltrs, Outputs: outputs}.Execute()
}

// init initializes the RunFns with a containerFilterProvider.
func (r *RunFns) init() {
	// if containerFilterProvider hasn't been set, use the default
	if r.containerFilterProvider == nil {
		r.containerFilterProvider = func(image, path string, api *yaml.RNode) kio.Filter {
			cf := &filters.ContainerFilter{Image: image, Config: api}
			cf.SetMountPath(filepath.Join(r.Path, path))
			return cf
		}
	}
}
