// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var flags struct {
	vars bool
}

// NewCmdFix returns an instance of 'fix' subcommand.
func NewCmdFix(fSys filesys.FileSystem, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Fix the missing fields in kustomization file",
		Long:  "",
		Example: `
	# Fix the missing and deprecated fields in kustomization file
	kustomize edit fix

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunFix(fSys, w)
		},
	}
	AddFlagVars(cmd.Flags())
	return cmd
}

// RunFix runs `fix` command
func RunFix(fSys filesys.FileSystem, w io.Writer) error {
	var oldOutput bytes.Buffer
	oldBuildCmd := build.NewCmdBuild(fSys, build.MakeHelp(konfig.ProgramName, "build"), &oldOutput)
	oldBuildCmd.RunE(oldBuildCmd, nil)

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

	if flags.vars {
		err = ConvertVarsToReplacements(fSys, m)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, `
Fixed fields:
  patchesJson6902 -> patches
  commonLabels -> labels
  vars -> replacements`)

	} else {
		fmt.Fprintln(w, `
Fixed fields:
  patchesJson6902 -> patches
  commonLabels -> labels

To convert vars -> replacements, run the command `+"`kustomize edit fix --vars`"+`

WARNING: Converting vars to replacements will potentially overwrite many resource files 
and the resulting files may not produce the same output when `+"`kustomize build`"+` is run. 
We recommend doing this in a clean git repository where the change is easy to undo.`)
	}

	writeErr := mf.Write(m)

	var fixedOutput bytes.Buffer
	fixedBuildCmd := build.NewCmdBuild(fSys, build.MakeHelp(konfig.ProgramName, "build"), &fixedOutput)
	err = fixedBuildCmd.RunE(fixedBuildCmd, nil)
	if err != nil {
		fmt.Fprintf(w, "Warning: 'Fixed' kustomization now produces the error when running `kustomize build`: %s", err.Error())
	} else if fixedOutput.String() != oldOutput.String() {
		fmt.Fprintf(w, "Warning: 'Fixed' kustomization now produces different output when running `kustomize build`:\n...%s...\n", fixedOutput.String())
	}

	return writeErr
}

func AddFlagVars(set *pflag.FlagSet) {
	set.BoolVar(
		&flags.vars,
		"vars",
		false, // default
		`If specified, kustomize will attempt to convert vars to replacements. 
We recommend doing this in a clean git repository where the change is easy to undo.`)
}
