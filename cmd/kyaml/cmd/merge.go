// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

func GetMergeRunner() *MergeRunner {
	r := &MergeRunner{}
	c := &cobra.Command{
		Use:   "merge [SOURCE_DIR...] [DESTINATION_DIR]",
		Short: "Merge Resource configuration files",
		Long: `Merge Resource configuration files

Merge reads Kubernetes Resource yaml configuration files from stdin or sources packages and write
the result to stdout or a destination package.

Resources are merged using the Resource [apiVersion, kind, name, namespace] as the key.  If any of
these are missing, merge will default the missing values to empty.

Resources specified later are high-precedence (the source) and Resources specified
earlier are lower-precedence (the destination).

For information on merge rules, run:

	kyaml help merge
`,
		Example: `cat resources_and_patches.yaml | kyaml merge > merged_resources.yaml`,
		RunE:    r.runE,
	}
	r.Command = c
	r.Command.Flags().BoolVar(&r.InvertOrder, "invert-order", false,
		"if true, merge Resources in the reverse order")
	return r
}

func MergeCommand() *cobra.Command {
	return GetMergeRunner().Command
}

// MergeRunner contains the run function
type MergeRunner struct {
	Command     *cobra.Command
	InvertOrder bool
}

func (r *MergeRunner) runE(c *cobra.Command, args []string) error {
	var inputs []kio.Reader
	// add the packages in reverse order -- the arg list should be highest precedence first
	// e.g. merge from -> to, but the MergeFilter is highest precedence last
	for i := len(args) - 1; i >= 0; i-- {
		inputs = append(inputs, kio.LocalPackageReader{PackagePath: args[i]})
	}
	// if there is no "from" package, read from stdin
	rw := &kio.ByteReadWriter{
		Reader:                c.InOrStdin(),
		Writer:                c.OutOrStdout(),
		KeepReaderAnnotations: true,
	}
	if len(inputs) < 2 {
		inputs = append(inputs, rw)
	}

	// write to the "to" package if specified
	var outputs []kio.Writer
	if len(args) != 0 {
		outputs = append(outputs, kio.LocalPackageWriter{PackagePath: args[len(args)-1]})
	}
	// if there is no "to" package, write to stdout
	if len(outputs) == 0 {
		outputs = append(outputs, rw)
	}

	filters := []kio.Filter{filters.MergeFilter{}, filters.FormatFilter{}}
	return kio.Pipeline{Inputs: inputs, Filters: filters, Outputs: outputs}.Execute()
}
