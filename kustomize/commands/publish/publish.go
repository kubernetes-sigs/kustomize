// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package publish

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/utils/clock"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/oci"
)

type publishOptions struct {
	registry  string
	createdAt time.Time
	noVerify  bool
}

// NewCmdEdit returns an instance of 'edit' subcommand.
func NewCmdPublish(
	fSys filesys.FileSystem,
	clock clock.PassiveClock,
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
			err := o.Validate(clock, args)
			if err != nil {
				return err
			}
			return o.RunPublish(fSys)
		},

		Args: cobra.MinimumNArgs(1),
	}

	// cmd.Flags().StringVar(
	// 	&o.createdAt,
	// 	"created-at",
	// 	"",
	// 	"The timestamp of the published artifact.  It must be supplied for reproducible builds.  Defaults to the current timestamp.",
	// )
	cmd.Flags().BoolVar(
		&o.noVerify,
		"no-verify",
		false,
		"skip validation for resources",
	)
	return cmd
}

// Validate validates addResource command.
func (o *publishOptions) Validate(clock clock.PassiveClock, args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a registry")
	}
	o.registry = args[0]

	o.createdAt = clock.Now()

	// if o.createdAt == "" {
	// 	o.createdAt = time.Now().Format(time.RFC3339)
	// } else {
	// 	parsed, err := time.Parse("", o.createdAt)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	o.createdAt = parsed.Format(time.RFC3339)
	// }
	return nil
}

// RunAddResource runs addResource command (do real work).
func (o *publishOptions) RunPublish(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	kustomization, err := mf.Read()
	if err != nil {
		return err
	}

	var dir string = filepath.Dir(mf.GetPath())

	fs, err := file.New("")
	if err != nil {
		return err
	}
	defer fs.Close()

	ctx := context.Background()
	fileDescriptor, err := fs.Add(ctx, ".", "", dir)
	if err != nil {
		return err
	}

	opts := oras.PackManifestOptions{
		Layers: []v1.Descriptor{
			fileDescriptor,
		},
		ManifestAnnotations: map[string]string{
			"org.opencontainers.image.created": o.createdAt.Format(time.RFC3339),
		},
	}

	artifactType := fmt.Sprintf("application/vnd.%s+%s", strings.ToLower(strings.ReplaceAll(kustomization.APIVersion, "/", ".")), strings.ToLower(kustomization.Kind))
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

	// }
	// reg, err := remote.NewRegistry(o.registry)
	// reg.PlainHTTP = true

	// dst, err := reg.Repository(ctx, "destination")
	// if err != nil {
	// 	panic(err) // Handle error
	// }

	desc, err := oras.Copy(ctx, fs, tag, dst, tag, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	log.Printf(`SUCCESS: published %s:%s@%s\n`, o.registry, tag, desc.Digest)

	return nil
}
