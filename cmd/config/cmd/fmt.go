// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

// FmtCmd returns a command FmtRunner.
func GetFmtRunner() *FmtRunner {
	r := &FmtRunner{}
	c := &cobra.Command{
		Use:   "fmt",
		Short: "Format yaml configuration files",
		Long: `Format yaml configuration files

Fmt will format input by ordering fields and unordered list items in Kubernetes
objects.  Inputs may be directories, files or stdin, and their contents must
include both apiVersion and kind fields.

- Stdin inputs are formatted and written to stdout
- File inputs (args) are formatted and written back to the file
- Directory inputs (args) are walked, each encountered .yaml and .yml file
  acts as an input

For inputs which contain multiple yaml documents separated by \n---\n,
each document will be formatted and written back to the file in the original
order.

Field ordering roughly follows the ordering defined in the source Kubernetes
resource definitions (i.e. go structures), falling back on lexicographical
sorting for unrecognized fields.

Unordered list item ordering is defined for specific Resource types and
field paths.

- .spec.template.spec.containers (by element name)
- .webhooks.rules.operations (by element value)
`,
		Example: `
	# format file1.yaml and file2.yml
	kyaml fmt file1.yaml file2.yml

	# format all *.yaml and *.yml recursively traversing directories
	kyaml fmt my-dir/

	# format kubectl output
	kubectl get -o yaml deployments | kyaml fmt

	# format kustomize output
	kustomize build | kyaml fmt
`,
		RunE:    r.runE,
		PreRunE: r.preRunE,
	}
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

func FmtCommand() *cobra.Command {
	return GetFmtRunner().Command
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
