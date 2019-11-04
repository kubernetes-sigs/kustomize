// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/util"
)

type createFlags struct {
	resources       string
	namespace       string
	annotations     string
	labels          string
	prefix          string
	suffix          string
	detectResources bool
	detectRecursive bool
	path            string
}

// NewCmdCreate returns an instance of 'create' subcommand.
func NewCmdCreate(fSys filesys.FileSystem, uf ifc.KunstructuredFactory) *cobra.Command {
	opts := createFlags{path: "."}
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a new kustomization in the current directory",
		Long:  "",
		Example: `
	# Create a new overlay from the base '../base".
	kustomize create --resources ../base

	# Create a new kustomization detecting resources in the current directory.
	kustomize create --autodetect

	# Create a new kustomization with multiple resources and fields set.
	kustomize create --resources deployment.yaml,service.yaml,../base --namespace staging --nameprefix acme-
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts, fSys, uf)
		},
	}
	c.Flags().StringVar(
		&opts.resources,
		"resources",
		"",
		"Name of a file containing a file to add to the kustomization file.")
	c.Flags().StringVar(
		&opts.namespace,
		"namespace",
		"",
		"Set the value of the namespace field in the customization file.")
	c.Flags().StringVar(
		&opts.annotations,
		"annotations",
		"",
		"Add one or more common annotations.")
	c.Flags().StringVar(
		&opts.labels,
		"labels",
		"",
		"Add one or more common labels.")
	c.Flags().StringVar(
		&opts.prefix,
		"nameprefix",
		"",
		"Sets the value of the namePrefix field in the kustomization file.")
	c.Flags().StringVar(
		&opts.suffix,
		"namesuffix",
		"",
		"Sets the value of the nameSuffix field in the kustomization file.")
	c.Flags().BoolVar(
		&opts.detectResources,
		"autodetect",
		false,
		"Search for kubernetes resources in the current directory to be added to the kustomization file.")
	c.Flags().BoolVar(
		&opts.detectRecursive,
		"recursive",
		false,
		"Enable recursive directory searching for resource auto-detection.")
	return c
}

func runCreate(opts createFlags, fSys filesys.FileSystem, uf ifc.KunstructuredFactory) error {
	var resources []string
	var err error
	if opts.resources != "" {
		resources, err = util.GlobPatterns(fSys, strings.Split(opts.resources, ","))
		if err != nil {
			return err
		}
	}
	if _, err = kustfile.NewKustomizationFile(fSys); err == nil {
		return fmt.Errorf("kustomization file already exists")
	}
	if opts.detectResources {
		detected, err := detectResources(fSys, uf, opts.path, opts.detectRecursive)
		if err != nil {
			return err
		}
		for _, resource := range detected {
			if kustfile.StringInSlice(resource, resources) {
				continue
			}
			resources = append(resources, resource)
		}
	}
	f, err := fSys.Create("kustomization.yaml")
	if err != nil {
		return err
	}
	f.Close()
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	m.Resources = resources
	m.Namespace = opts.namespace
	m.NamePrefix = opts.prefix
	m.NameSuffix = opts.suffix
	annotations, err := util.ConvertToMap(opts.annotations, "annotation")
	if err != nil {
		return err
	}
	m.CommonAnnotations = annotations
	labels, err := util.ConvertToMap(opts.labels, "label")
	if err != nil {
		return err
	}
	m.CommonLabels = labels
	return mf.Write(m)
}

func detectResources(fSys filesys.FileSystem, uf ifc.KunstructuredFactory, base string, recursive bool) ([]string, error) {
	var paths []string
	err := fSys.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == base {
			return nil
		}
		if info.IsDir() {
			if !recursive {
				return filepath.SkipDir
			}
			// If a sub-directory contains an existing kustomization file add the
			// directory as a resource and do not decend into it.
			for _, kfilename := range konfig.RecognizedKustomizationFileNames() {
				if fSys.Exists(filepath.Join(path, kfilename)) {
					paths = append(paths, path)
					return filepath.SkipDir
				}
			}
			return nil
		}
		fContents, err := fSys.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := uf.SliceFromBytes(fContents); err != nil {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths, err
}
