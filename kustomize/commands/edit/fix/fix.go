// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
)

// NewCmdFix returns an instance of 'fix' subcommand.
func NewCmdFix(fSys filesys.FileSystem) *cobra.Command {
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
func RunFix(fSys filesys.FileSystem) error {
	tmpdir := fSys.DeepCopy()
	if err := DoFix(tmpdir); err != nil {
		return err
	}

	// compare output of `kustomize build`
	var fixedOutput bytes.Buffer
	fixedBuildCmd := build.NewCmdBuild(tmpdir, build.MakeHelp(konfig.ProgramName, "build"), &fixedOutput)
	fixedBuildCmd.RunE(fixedBuildCmd, nil)

	var origOutput bytes.Buffer
	origBuildCmd := build.NewCmdBuild(fSys, build.MakeHelp(konfig.ProgramName, "build"), &origOutput)
	origBuildCmd.RunE(origBuildCmd, nil)

	if fixedOutput.String() != origOutput.String() {
		return fmt.Errorf("could not automatically fix kustomization file, build output differs")
	}

	return DoFix(fSys)
}

// DoFix does the actual fixing work of fixing the missing and
// deprecated fields in the kustomization file
func DoFix(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	err = m.FixKustomizationPreMarshalling()
	if err != nil {
		return err
	}
	return mf.Write(m)
}
