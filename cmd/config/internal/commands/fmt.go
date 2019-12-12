// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

// FmtCmd returns a command FmtRunner.
func GetFmtRunner(name string) *FmtRunner {
	r := &FmtRunner{}
	c := &cobra.Command{
		Use:     "fmt",
		Short:   commands.FmtShort,
		Long:    commands.FmtLong,
		Example: commands.FmtExamples,
		RunE:    r.runE,
		PreRunE: r.preRunE,
	}
	fixDocs(name, c)
	c.Flags().StringVar(&r.FilenamePattern, "pattern", filters.DefaultFilenamePattern,
		`pattern to use for generating filenames for resources -- may contain the following
formatting substitution verbs {'%n': 'metadata.name', '%s': 'metadata.namespace', '%k': 'kind'}`)
	c.Flags().BoolVar(&r.SetFilenames, "set-filenames", false,
		`if true, set default filenames on Resources without them`)
	c.Flags().BoolVar(&r.KeepAnnotations, "keep-annotations", false,
		`if true, keep index and filename annotations set on Resources.`)
	c.Flags().BoolVar(&r.Override, "override", false,
		`if true, override existing filepath annotations.`)
	r.Command = c
	return r
}

func FmtCommand(name string) *cobra.Command {
	return GetFmtRunner(name).Command
}

// FmtRunner contains the run function
type FmtRunner struct {
	Command         *cobra.Command
	FilenamePattern string
	SetFilenames    bool
	KeepAnnotations bool
	Override        bool
}

func (r *FmtRunner) preRunE(c *cobra.Command, args []string) error {
	if r.SetFilenames {
		r.KeepAnnotations = true
	}
	return nil
}

func (r *FmtRunner) runE(c *cobra.Command, args []string) error {
	f := []kio.Filter{filters.FormatFilter{}}

	// format with file names
	if r.SetFilenames {
		f = append(f, &filters.FileSetter{
			FilenamePattern: r.FilenamePattern,
			Override:        r.Override,
		})
	}

	// format stdin if there are no args
	if len(args) == 0 {
		rw := &kio.ByteReadWriter{
			Reader:                c.InOrStdin(),
			Writer:                c.OutOrStdout(),
			KeepReaderAnnotations: r.KeepAnnotations,
		}
		return handleError(c, kio.Pipeline{
			Inputs: []kio.Reader{rw}, Filters: f, Outputs: []kio.Writer{rw}}.Execute())
	}

	for i := range args {
		path := args[i]
		rw := &kio.LocalPackageReadWriter{
			NoDeleteFiles:         true,
			PackagePath:           path,
			KeepReaderAnnotations: r.KeepAnnotations}
		err := kio.Pipeline{
			Inputs: []kio.Reader{rw}, Filters: f, Outputs: []kio.Writer{rw}}.Execute()
		if err != nil {
			return handleError(c, err)
		}
	}
	return nil
}
