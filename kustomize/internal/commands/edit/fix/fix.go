// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

// NewCmdFix returns an instance of 'fix' subcommand.
func NewCmdFix(fSys fs.FileSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Fix the missing fields in kustomization file",
		Long:  "",
		Example: `
	# Fix the missing and deprecated fields in kustomization file
	kustomize edit fix

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunFix(fSys)
		},
	}
	return cmd
}

// RunFix runs `fix` command
func RunFix(fSys fs.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	return mf.Write(m)
}
