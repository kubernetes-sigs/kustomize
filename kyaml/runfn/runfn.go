// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package runfn

import (
	"io"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// RunFns runs the set of configuration functions in a local directory against
// the Resources in that directory
type RunFns struct {
	StorageMounts []filters.StorageMount

	// Path is the path to the directory containing functions
	Path string

	// FunctionPaths Paths allows functions to be specified outside the configuration
	// directory.
	// Functions provided on FunctionPaths are globally scoped.
	FunctionPaths []string

	// GlobalScope if true, functions read from input will be scoped globally rather
	// than only to Resources under their subdirs.
	GlobalScope bool

	// Output can be set to write the result to Output rather than back to the directory
	Output io.Writer

	// NoFunctionsFromInput if set to true will not read any functions from the input,
	// and only use explicit sources
	NoFunctionsFromInput *bool

	// for testing purposes only
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

	fltrs, err := r.getFilters()
	if err != nil {
		return err
	}

	return r.runFunctions(fltrs)
}

func (r RunFns) getFilters() ([]kio.Filter, error) {
	var fltrs []kio.Filter

	// implicit filters from the input Resources
	f, err := r.getFunctionsFromInput()
	if err != nil {
		return nil, err
	}
	fltrs = append(fltrs, f...)

	// explicit filters from a list of directories
	f, err = r.getFunctionsFromDirList()
	if err != nil {
		return nil, err
	}
	fltrs = append(fltrs, f...)
	return fltrs, nil
}

// runFunctions runs the fltrs against the input
func (r RunFns) runFunctions(fltrs []kio.Filter) error {
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

// getFunctionsFromInput scans the input for functions and runs them
func (r RunFns) getFunctionsFromInput() ([]kio.Filter, error) {
	if *r.NoFunctionsFromInput {
		return nil, nil
	}

	var fltrs []kio.Filter
	buff := &kio.PackageBuffer{}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{kio.LocalPackageReader{PackagePath: r.Path}},
		Filters: []kio.Filter{&filters.IsReconcilerFilter{}},
		Outputs: []kio.Writer{buff},
	}.Execute()
	if err != nil {
		return nil, err
	}
	sortFns(buff)
	for i := range buff.Nodes {
		api := buff.Nodes[i]
		img, path := filters.GetContainerName(api)
		fltrs = append(fltrs, r.containerFilterProvider(img, path, api))
	}
	return fltrs, nil
}

// getFunctionsFromDirList returns the set of functions read from r.FunctionPaths
// as a slice of Filters
func (r RunFns) getFunctionsFromDirList() ([]kio.Filter, error) {
	var fltrs []kio.Filter
	buff := &kio.PackageBuffer{}
	for i := range r.FunctionPaths {
		err := kio.Pipeline{
			Inputs:  []kio.Reader{kio.LocalPackageReader{PackagePath: r.FunctionPaths[i]}},
			Outputs: []kio.Writer{buff},
		}.Execute()
		if err != nil {
			return nil, err
		}
	}
	for i := range buff.Nodes {
		api := buff.Nodes[i]
		img, path := filters.GetContainerName(api)
		c := r.containerFilterProvider(img, path, api)
		cf, ok := c.(*filters.ContainerFilter)
		if ok {
			// functions provided on FunctionPaths are globally scoped
			cf.GlobalScope = true
		}
		fltrs = append(fltrs, c)
	}
	return fltrs, nil
}

// sortFns sorts functions so that functions with the longest paths come first
func sortFns(buff *kio.PackageBuffer) {
	// sort the nodes so that we traverse them depth first
	// functions deeper in the file system tree should be run first
	sort.Slice(buff.Nodes, func(i, j int) bool {
		mi, _ := buff.Nodes[i].GetMeta()
		pi := mi.Annotations[kioutil.PathAnnotation]
		if path.Base(path.Dir(pi)) == "functions" {
			// don't count the functions dir, the functions are scoped 1 level above
			pi = path.Dir(path.Dir(pi))
		} else {
			pi = path.Dir(pi)
		}

		mj, _ := buff.Nodes[j].GetMeta()
		pj := mj.Annotations[kioutil.PathAnnotation]
		if path.Base(path.Dir(pj)) == "functions" {
			// don't count the functions dir, the functions are scoped 1 level above
			pj = path.Dir(path.Dir(pj))
		} else {
			pj = path.Dir(pj)
		}

		// i is "less" than j (comes earlier) if its depth is greater -- e.g. run
		// i before j if it is deeper in the directory structure
		li := len(strings.Split(pi, "/"))
		if pi == "." {
			// local dir should have 0 path elements instead of 1
			li = 0
		}
		lj := len(strings.Split(pj, "/"))
		if pj == "." {
			// local dir should have 0 path elements instead of 1
			lj = 0
		}
		if li != lj {
			// use greater-than because we want to sort with the longest
			// paths FIRST rather than last
			return li > lj
		}

		// sort by path names if depths are equal
		return pi < pj
	})
}

// init initializes the RunFns with a containerFilterProvider.
func (r *RunFns) init() {
	if r.NoFunctionsFromInput == nil {
		nfn := len(r.FunctionPaths) > 0
		r.NoFunctionsFromInput = &nfn
	}

	// if containerFilterProvider hasn't been set, use the default
	if r.containerFilterProvider == nil {
		r.containerFilterProvider = func(image, path string, api *yaml.RNode) kio.Filter {
			cf := &filters.ContainerFilter{
				Image:         image,
				Config:        api,
				StorageMounts: r.StorageMounts,
				GlobalScope:   r.GlobalScope,
			}
			return cf
		}
	}
}
