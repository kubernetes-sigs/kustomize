// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package publish

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/oci"
)

type publishOptions struct {
	registry string
	noVerify bool
}

// NewCmdEdit returns an instance of 'edit' subcommand.
func NewCmdPublish(
	fSys filesys.FileSystem, v ifc.Validator, rf *resource.Factory,
) *cobra.Command {
	var o publishOptions

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publishes a kustomization resource to an OCI registry",
		Long:  "",
		Example: `
		publish <registry>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunPublish(fSys)
		},

		Args: cobra.MinimumNArgs(1),
	}
	cmd.Flags().BoolVar(&o.noVerify, "no-verify", false,
		"skip validation for resources",
	)
	return cmd
}

// Validate validates addResource command.
func (o *publishOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a registry")
	}
	o.registry = args[0]
	return nil
}

// RunAddResource runs addResource command (do real work).
func (o *publishOptions) RunPublish(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	if _, err = mf.Read(); err != nil {
		return err
	}

	var dir string = filepath.Dir(mf.GetPath())

	fs, err := file.New(dir)
	if err != nil {
		return err
	}
	defer fs.Close()

	ctx := context.Background()
	fileDescriptors := make([]v1.Descriptor, 0)

	if err = fSys.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		var mediaType string = "application/vnd.kustomize.unknown.v1beta1"
		var b string = filepath.Base(path)
		for _, kfilename := range konfig.RecognizedKustomizationFileNames() {
			if b == kfilename {
				mediaType = "application/vnd.kustomize.config.k8s.io.v1beta1+yaml"
				break
			}
		}

		fileDescriptor, err := fs.Add(ctx, path, mediaType, "")
		if err != nil {
			return err
		}
		fileDescriptors = append(fileDescriptors, fileDescriptor)

		return nil
	}); err != nil {
		return err
	}

	artifactType := "application/vnd.kustomize.artifact"
	opts := oras.PackManifestOptions{
		Layers: fileDescriptors,
	}
	manifestDescriptor, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, artifactType, opts)
	if err != nil {
		return err
	}

	tag := "latest"
	if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
		return err
	}

	dst, err := oci.New(o.registry)
	if err != nil {
		return err
	}

	desc, err := oras.Copy(ctx, fs, tag, dst, tag, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	log.Printf("SUCCESS: copied %s desc %q\n", dir, desc)

	return nil
}
