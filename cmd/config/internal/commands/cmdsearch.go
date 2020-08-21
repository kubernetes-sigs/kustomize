// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/setters2"
)

// NewSearchRunner returns a command SearchRunner.
func NewSearchRunner(name string) *SearchRunner {
	r := &SearchRunner{}
	c := &cobra.Command{
		Use:     "search DIR",
		RunE:    r.runE,
		PreRunE: r.preRunE,
		Args:    cobra.ExactArgs(1),
	}
	fixDocs(name, c)
	c.Flags().StringVar(&r.SearchReplace.Value, "by-value", "",
		"Match by value of a field.")
	c.Flags().StringVar(&r.SearchReplace.ValueRegex, "by-value-regex", "",
		"Replace the value of the matching fields with the given literal value.")
	c.Flags().StringVar(&r.SearchReplace.ReplaceLiteral, "replace-with-literal", "",
		"Match by Regex for the value of a field.")

	r.Command = c
	return r
}

func SearchCommand(name string) *cobra.Command {
	return NewSearchRunner(name).Command
}

// SearchRunner contains the SearchReplace function
type SearchRunner struct {
	Command       *cobra.Command
	SearchReplace setters2.SearchReplace
}

func (r *SearchRunner) preRunE(c *cobra.Command, args []string) error {
	if !c.Flag("by-value").Changed &&
		!c.Flag("by-value-regex").Changed &&
		c.Flag("replace-with-literal").Changed {
		return errors.Errorf("replace-with-literal can only be invoked along with by-value or by-value-regex")
	}
	return nil
}

func (r *SearchRunner) runE(c *cobra.Command, args []string) error {
	err := r.SearchReplace.Perform(args[0])
	fmt.Fprintf(c.OutOrStdout(), "matched %d fields\n", r.SearchReplace.Count)
	for _, node := range r.SearchReplace.Match {
		res, err := node.String()
		if err != nil {
			return errors.Wrap(err)
		}
		fmt.Fprintf(c.OutOrStdout(), "---\n")
		fmt.Fprint(c.OutOrStdout(), res)
	}
	return errors.Wrap(err)
}
