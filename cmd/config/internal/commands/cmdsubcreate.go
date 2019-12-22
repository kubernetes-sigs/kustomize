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
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/set"
)

// NewSubSetRunner returns a command runner.
func NewSubSetRunner(parent string) *SubSetRunner {
	r := &SubSetRunner{}
	set := &cobra.Command{
		Use:     "create PKG_DIR NAME [VALUE]",
		Args:    cobra.ExactArgs(3),
		Short:   commands.SubsetShort,
		Long:    commands.SubsetLong,
		Example: commands.SubsetExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
	}
	set.Flags().StringVar(&r.Set.Marker.OwnedBy, "owned-by", "",
		"set this owner on for the current value.")
	set.Flags().StringVar(&r.Set.Marker.Description, "description", "",
		"set this description for the current value description.")
	set.Flags().StringVar(&r.Set.Marker.Substitution.Marker, "marker", "[MARKER]",
		"use this marker.")
	set.Flags().StringVar(&r.Set.Marker.Field, "field", "",
		"name of the field to set -- e.g. --field port")
	set.Flags().StringVar(&r.Set.ResourceMeta.Name, "name", "",
		"name of the Resource on which to set the substitution.")
	set.Flags().StringVar(&r.Set.ResourceMeta.Kind, "kind", "",
		"kind of the Resource on which to set substitution.")
	set.Flags().StringVar(&r.Set.Marker.Type, "type", "",
		"field type -- e.g. int,float,bool,string.")
	set.Flags().BoolVar(&r.Set.Marker.PartialMatch, "substring", false,
		"if true, the value may be a substring of the current value.")
	fixDocs(parent, set)
	set.MarkFlagRequired("type")
	set.MarkFlagRequired("field")
	r.Command = set
	return r
}

func SubSetCommand(parent string) *cobra.Command {
	return NewSubSetRunner(parent).Command
}

type SubSetRunner struct {
	Command *cobra.Command
	Set     set.SetSubstitutionMarker
}

func (r *SubSetRunner) runE(c *cobra.Command, args []string) error {
	return handleError(c, r.set(c, args))
}

func (r *SubSetRunner) preRunE(c *cobra.Command, args []string) error {
	r.Set.Marker.Substitution.Name = args[1]
	r.Set.Marker.Substitution.Value = args[2]
	return nil
}

// perform the substitutions
func (r *SubSetRunner) set(c *cobra.Command, args []string) error {
	rw := &kio.LocalPackageReadWriter{
		PackagePath: args[0],
	}
	// add the substitution marker to the Resource
	err := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{&r.Set},
		Outputs: []kio.Writer{rw},
	}.Execute()
	if err != nil {
		return err
	}
	return nil
}
