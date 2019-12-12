// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func GetCountRunner(name string) *CountRunner {
	r := &CountRunner{}
	c := &cobra.Command{
		Use:     "count DIR...",
		Short:   commands.CountShort,
		Long:    commands.CountLong,
		Example: commands.CountExamples,
		RunE:    r.runE,
	}
	fixDocs(name, c)
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	c.Flags().BoolVar(&r.Kind, "kind", true,
		"count resources by kind.")

	r.Command = c
	return r
}

func CountCommand(name string) *cobra.Command {
	return GetCountRunner(name).Command
}

// CountRunner contains the run function
type CountRunner struct {
	IncludeSubpackages bool
	Kind               bool
	Command            *cobra.Command
}

func (r *CountRunner) runE(c *cobra.Command, args []string) error {
	var inputs []kio.Reader
	for _, a := range args {
		inputs = append(inputs, kio.LocalPackageReader{
			PackagePath:        a,
			IncludeSubpackages: r.IncludeSubpackages,
		})
	}
	if len(inputs) == 0 {
		inputs = append(inputs, &kio.ByteReader{Reader: c.InOrStdin()})
	}

	var out []kio.Writer
	if r.Kind {
		out = append(out, kio.WriterFunc(func(nodes []*yaml.RNode) error {
			count := map[string]int{}
			k := sets.String{}
			for _, n := range nodes {
				m, _ := n.GetMeta()
				count[m.Kind]++
				k.Insert(m.Kind)
			}
			order := k.List()
			sort.Strings(order)
			for _, k := range order {
				fmt.Fprintf(c.OutOrStdout(), "%s: %d\n", k, count[k])
			}

			return nil
		}))

	} else {
		out = append(out, kio.WriterFunc(func(nodes []*yaml.RNode) error {
			fmt.Fprintf(c.OutOrStdout(), "%d\n", len(nodes))
			return nil
		}))
	}
	return handleError(c, kio.Pipeline{
		Inputs:  inputs,
		Outputs: out,
	}.Execute())
}
