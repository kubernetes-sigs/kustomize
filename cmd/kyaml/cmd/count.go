// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func GetCountRunner() *CountRunner {
	r := &CountRunner{}
	c := &cobra.Command{
		Use:   "count DIR...",
		Short: "Count Resources Config from a local directory",
		Long: `Count Resources Config from a local directory.

  DIR:
    Path to local directory.
`,
		Example: `# print Resource counts from a directory
kyaml count my-dir/
`,
		RunE: r.runE,
	}
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	c.Flags().BoolVar(&r.Kind, "kind", true,
		"count resources by kind.")

	r.Command = c
	return r
}

func CountCommand() *cobra.Command {
	return GetCountRunner().Command
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
