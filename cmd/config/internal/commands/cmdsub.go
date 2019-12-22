// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/set"
)

// NewSubRunner returns a command runner.
func NewSubRunner(parent string) *SubRunner {
	r := &SubRunner{}
	c := &cobra.Command{
		Use:     "set DIR [NAME] [VALUE]",
		Args:    cobra.RangeArgs(1, 3),
		Short:   commands.SubShort,
		Long:    commands.SubLong,
		Example: commands.SubExamples,
		Aliases: []string{"sub"},
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	c.Flags().BoolVar(&r.Perform.Override, "override", true,
		"override previously substituted values.")
	c.Flags().BoolVar(&r.Perform.Revert, "revert", false,
		"override previously substituted values.")
	fixDocs(parent, c)
	r.Command = c
	c.AddCommand(SubSetCommand(parent))
	return r
}

func SubCommand(parent string) *cobra.Command {
	return NewSubRunner(parent).Command
}

type SubRunner struct {
	Command *cobra.Command
	Lookup  set.LookupSubstitutions
	Perform set.PerformSubstitutions
}

func (r *SubRunner) preRunE(c *cobra.Command, args []string) error {
	if len(args) > 1 {
		r.Perform.Name = args[1]
		r.Lookup.Name = args[1]
	}
	if len(args) > 2 {
		r.Perform.NewValue = args[2]
	}
	if len(args) < 2 && r.Perform.Revert {
		return errors.Errorf("must specify NAME with --revert")
	}

	var mutex int
	if r.Perform.Revert {
		mutex++
	}
	if r.Perform.Override {
		mutex++
	}
	if mutex > 1 {
		return errors.Errorf("--revert, --override are mutually exclusive")
	}

	return nil
}

func (r *SubRunner) runE(c *cobra.Command, args []string) error {

	if len(args) == 3 {
		return handleError(c, r.perform(c, args))
	}
	if len(args) == 2 && r.Perform.Revert {
		return handleError(c, r.perform(c, args))
	}

	return handleError(c, r.lookup(c, args))
}

func (r *SubRunner) lookup(c *cobra.Command, args []string) error {
	// lookup the substitutions
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.LocalPackageReader{PackagePath: args[0]}},
		Filters: []kio.Filter{&r.Lookup},
	}.Execute()
	if err != nil {
		return err
	}

	remaining := false
	table := tablewriter.NewWriter(c.OutOrStdout())
	table.SetRowLine(false)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator(" ")
	table.SetCenterSeparator(" ")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{
		"NAME", "DESCRIPTION", "VALUE", "TYPE", "COUNT", "SUBSTITUTED", "OWNER",
	})
	for i := range r.Lookup.SubstitutionCounts {
		s := r.Lookup.SubstitutionCounts[i]
		remaining = remaining || s.Count > s.CountComplete
		v := s.CurrentValue
		if s.CurrentValue == "" {
			v = s.Marker
		}
		table.Append([]string{
			s.Name,
			"'" + s.Description + "'",
			v,
			fmt.Sprintf("%v", s.Type),
			fmt.Sprintf("%d", s.Count),
			fmt.Sprintf("%v", s.Count == s.CountComplete),
			s.OwnedBy,
		})
	}
	table.Render()

	if remaining {
		os.Exit(1)
	}
	return nil
}

// perform the substitutions
func (r *SubRunner) perform(c *cobra.Command, args []string) error {
	rw := &kio.LocalPackageReadWriter{
		PackagePath: args[0],
	}
	// perform the substitutions in the package
	err := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{&r.Perform},
		Outputs: []kio.Writer{rw},
	}.Execute()
	if err != nil {
		return err
	}
	fmt.Fprintf(c.OutOrStdout(), "performed %d substitutions\n", r.Perform.Count)
	return nil
}
