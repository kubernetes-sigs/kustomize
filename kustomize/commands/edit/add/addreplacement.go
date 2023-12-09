// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type addReplacementOptions struct {
	Replacement types.ReplacementField
}

func newCmdAddReplacement(fSys filesys.FileSystem) *cobra.Command {
	var o addReplacementOptions
	cmd := &cobra.Command{
		Use:   "replacement",
		Short: "add an item to replacement field",
		Long: `this command will add an item to  replacement field in the kustomization file.
The item must be either a file, or an inline string.
`,
		Example: `
	# Adds a replacement file to the kustomization file
	kustomize edit add replacement --path {filepath}
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}
			return o.RunAddReplacement(fSys)
		},
	}

	cmd.Flags().StringVar(&o.Replacement.Path, "path", "", "Path to the replacement file.")
	return cmd
}

// Validate validate add replacement command
func (o *addReplacementOptions) Validate() error {
	if o.Replacement.Path == "" {
		return errors.New("must provide path to add replacement")
	}
	return nil
}

// RunAddReplacement runs addReplacement command
func (o *addReplacementOptions) RunAddReplacement(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return fmt.Errorf("failed to load kustomization file: %w", err)
	}

	m, err := mf.Read()
	if err != nil {
		return fmt.Errorf("failed to read kustomization file: %w", err)
	}

	for _, r := range m.Replacements {
		if len(r.Path) > 0 && r.Path == o.Replacement.Path {
			return fmt.Errorf("replacement for path %q already in %s file", r.Path, konfig.DefaultKustomizationFileName())
		}
	}
	m.Replacements = append(m.Replacements, o.Replacement)

	err = mf.Write(m)
	if err != nil {
		return fmt.Errorf("failed to write kustomization file: %w", err)
	}

	return nil
}
