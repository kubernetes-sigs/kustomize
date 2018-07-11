/*
Copyright 2017 The Kubernetes Authors.

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
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"errors"

	"github.com/kubernetes-sigs/kustomize/pkg/app"
	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
)

type buildOptions struct {
	kustomizationPath       string
	defaultRenamingBehavior string
}

// newCmdBuild creates a new build command.
func newCmdBuild(out io.Writer, fs fs.FileSystem) *cobra.Command {
	var o buildOptions

	cmd := &cobra.Command{
		Use:   "build [path]",
		Short: "Print current configuration per contents of " + constants.KustomizationFileName,
		Example: "Use the file somedir/" + constants.KustomizationFileName +
			" to generate a set of api resources:\nbuild somedir/",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunBuild(out, fs)
		},
	}

	cmd.Flags().StringVar(
		&o.defaultRenamingBehavior,
		"default-rename-behavior",
		"hash",
		"The default renaming behavior during the processing of configmaps and secrets. Can be 'none' or 'hash'.")

	return cmd
}

// Validate validates build command.
func (o *buildOptions) Validate(args []string) error {
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

// RunBuild runs build command.
func (o *buildOptions) RunBuild(out io.Writer, fs fs.FileSystem) error {
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

	allResources, err := application.MakeCustomizedResMap(resource.NewRenamingBehavior(o.defaultRenamingBehavior))

	if err != nil {
		return err
	}

	// Output the objects.
	res, err := allResources.EncodeAsYaml()
	if err != nil {
		return err
	}
	_, err = out.Write(res)
	return err
}
