// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

func GetMerge3Runner(name string) *Merge3Runner {
	r := &Merge3Runner{}
	c := &cobra.Command{
		Use:     "merge3 --ancestor [ORIGINAL_DIR] --from [UPDATED_DIR] --to [DESTINATION_DIR]",
		Short:   commands.Merge3Short,
		Long:    commands.Merge3Long,
		Example: commands.Merge3Examples,
		RunE:    r.runE,
		Deprecated: "this will no longer be available in kustomize v5.\n" +
			"See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.",
	}
	runner.FixDocs(name, c)
	c.Flags().StringVar(&r.ancestor, "ancestor", "",
		"Path to original package")
	c.Flags().StringVar(&r.fromDir, "from", "",
		"Path to updated package")
	c.Flags().StringVar(&r.toDir, "to", "",
		"Path to destination package")
	c.Flags().BoolVar(&r.path, "path-merge-key", false,
		"Use the path as part of the merge key when merging resources")

	r.Command = c
	return r
}

func Merge3Command(name string) *cobra.Command {
	return GetMerge3Runner(name).Command
}

// Merge3Runner contains the run function
type Merge3Runner struct {
	Command  *cobra.Command
	ancestor string
	fromDir  string
	toDir    string
	path     bool
}

func (r *Merge3Runner) runE(_ *cobra.Command, _ []string) error {
	matcher := filters.DefaultGVKNNMatcher{MergeOnPath: r.path}
	err := filters.Merge3{
		OriginalPath: r.ancestor,
		UpdatedPath:  r.fromDir,
		DestPath:     r.toDir,
		Matcher:      &matcher,
	}.Merge()
	if err != nil {
		return err
	}
	return nil
}
