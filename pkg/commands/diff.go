/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/kubernetes-sigs/kustomize/pkg/app"
	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/diff"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
)

type diffOptions struct {
	kustomizationPath       string
	defaultRenamingBehavior string
}

// newCmdDiff makes the diff command.
func newCmdDiff(out, errOut io.Writer, fs fs.FileSystem) *cobra.Command {
	var o diffOptions

	cmd := &cobra.Command{
		Use:   "diff [path]",
		Short: "diff between customized resources and uncustomized resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunDiff(out, errOut, fs)
		},
	}

	cmd.Flags().StringVar(
		&o.defaultRenamingBehavior,
		"default-rename-behavior",
		"hash",
		"The default renaming behavior during the processing of configmaps and secrets. Can be 'none' or 'hash'.")

	return cmd
}

// Validate validates diff command.
func (o *diffOptions) Validate(args []string) error {
	if len(args) > 1 {
		return errors.New("specify one path to " + constants.KustomizationFileName)
	}
	if len(args) == 0 {
		o.kustomizationPath = "./"
		return nil
	}
	o.kustomizationPath = args[0]
	return nil
}

// RunDiff gets the differences between Application.MakeCustomizedResMap() and Application.MakeUncustomizedResMap().
func (o *diffOptions) RunDiff(out, errOut io.Writer, fs fs.FileSystem) error {

	l := loader.Init([]loader.SchemeLoader{loader.NewFileLoader(fs)})

	absPath, err := filepath.Abs(o.kustomizationPath)
	if err != nil {
		return err
	}

	rootLoader, err := l.New(absPath)
	if err != nil {
		return err
	}

	application, err := app.NewApplication(rootLoader)
	if err != nil {
		return err
	}
	transformedResources, err := application.MakeCustomizedResMap(resource.NewRenamingBehavior(o.defaultRenamingBehavior))
	if err != nil {
		return err
	}
	rawResources, err := application.MakeUncustomizedResMap(resource.NewRenamingBehavior(o.defaultRenamingBehavior))
	if err != nil {
		return err
	}

	return diff.RunDiff(rawResources, transformedResources, out, errOut)
}
