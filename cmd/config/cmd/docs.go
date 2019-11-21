// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetDocsRunner() *DocsRunner {
	r := &DocsRunner{}
	c := &cobra.Command{
		Use:   "docs [API_TYPE]",
		Short: "Print out documentation for API resource",
		Long: `Print out documentation for API resource

The Docs command reads a JSON schema from the Kustomize API and outputs a pretiffied version of the documentation.

TODO:
<INSERT MORE DOCUMENTATION HERE>

For information on merge rules, run:

	kyaml help docs
`,
		Example: `kyaml docs kustomization`,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	r.Command = c
	r.Command.Flags().StringVar(&r.apiType, "api-type", "",
		"API type to print out")
	return r
}

// DocsCommand ...
func DocsCommand() *cobra.Command {
	return GetDocsRunner().Command
}

// DocsRunner contains the run function
type DocsRunner struct {
	Command *cobra.Command
	apiType string
}

func (r *DocsRunner) preRunE(c *cobra.Command, args []string) error {
	fmt.Println("Docs pre run. :)")
	return nil
}

func (r *DocsRunner) runE(c *cobra.Command, args []string) error {
	fmt.Println("Docs actual run. :-)")
	fmt.Println(args[0])
	return nil
}
