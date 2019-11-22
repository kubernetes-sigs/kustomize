// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/cmddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/api/resource"
)

// Cmd returns a command GrepRunner.
func GetGrepRunner(name string) *GrepRunner {
	r := &GrepRunner{}
	c := &cobra.Command{
		Use:     "grep QUERY [DIR]...",
		Short:   commands.GrepShort,
		Long:    commands.GrepLong,
		Example: commands.GrepExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
		Args:    cobra.MinimumNArgs(1),
	}
	fixDocs(name, c)
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	c.Flags().BoolVar(&r.KeepAnnotations, "annotate", true,
		"annotate resources with their file origins.")
	c.Flags().BoolVarP(&r.InvertMatch, "invert-match", "v", false,
		" Selected Resources are those not matching any of the specified patterns..")

	r.Command = c
	return r
}

func GrepCommand(name string) *cobra.Command {
	return GetGrepRunner(name).Command
}

// GrepRunner contains the run function
type GrepRunner struct {
	IncludeSubpackages bool
	KeepAnnotations    bool
	Command            *cobra.Command
	filters.GrepFilter
	Format bool
}

func (r *GrepRunner) preRunE(c *cobra.Command, args []string) error {
	r.GrepFilter.Compare = func(a, b string) (int, error) {
		qa, err := resource.ParseQuantity(a)
		if err != nil {
			return 0, fmt.Errorf("%s: %v", a, err)
		}
		qb, err := resource.ParseQuantity(b)
		if err != nil {
			return 0, err
		}

		return qa.Cmp(qb), err
	}
	parts, err := parseFieldPath(args[0])
	if err != nil {
		return err
	}

	var last []string
	if strings.Contains(parts[len(parts)-1], ">=") {
		last = strings.Split(parts[len(parts)-1], ">=")
		r.MatchType = filters.GreaterThanEq
	} else if strings.Contains(parts[len(parts)-1], "<=") {
		last = strings.Split(parts[len(parts)-1], "<=")
		r.MatchType = filters.LessThanEq
	} else if strings.Contains(parts[len(parts)-1], ">") {
		last = strings.Split(parts[len(parts)-1], ">")
		r.MatchType = filters.GreaterThan
	} else if strings.Contains(parts[len(parts)-1], "<") {
		last = strings.Split(parts[len(parts)-1], "<")
		r.MatchType = filters.LessThan
	} else {
		last = strings.Split(parts[len(parts)-1], "=")
		r.MatchType = filters.Regexp
	}
	if len(last) > 2 {
		return fmt.Errorf(
			"ambiguous match -- multiple of ['<', '>', '<=', '>=', '=' in final path element: %s",
			parts[len(parts)-1])
	}

	if len(last) > 1 {
		r.Value = last[1]
	}

	r.Path = append(parts[:len(parts)-1], last[0])
	return nil
}

func (r *GrepRunner) runE(c *cobra.Command, args []string) error {
	var filters = []kio.Filter{r.GrepFilter}

	var inputs []kio.Reader
	for _, a := range args[1:] {
		inputs = append(inputs, kio.LocalPackageReader{
			PackagePath:        a,
			IncludeSubpackages: r.IncludeSubpackages,
		})
	}
	if len(inputs) == 0 {
		inputs = append(inputs, &kio.ByteReader{Reader: c.InOrStdin()})
	}

	return handleError(c, kio.Pipeline{
		Inputs:  inputs,
		Filters: filters,
		Outputs: []kio.Writer{kio.ByteWriter{
			Writer:                c.OutOrStdout(),
			KeepReaderAnnotations: r.KeepAnnotations,
		}},
	}.Execute())
}
