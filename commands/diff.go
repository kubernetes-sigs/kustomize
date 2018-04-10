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

	"k8s.io/kubectl/pkg/kustomize/app"
	"k8s.io/kubectl/pkg/kustomize/util"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
	"k8s.io/kubectl/pkg/loader"
	"k8s.io/utils/exec"
)

type diffOptions struct {
	manifestPath string
}

// newCmdDiff makes the diff command.
func newCmdDiff(out, errOut io.Writer, fs fs.FileSystem) *cobra.Command {
	var o diffOptions

	cmd := &cobra.Command{
		Use:     "diff",
		Short:   "diff between transformed resources and untransformed resources",
		Long:    "diff between transformed resources and untransformed resources and the subpackages are all transformed.",
		Example: `diff -f .`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(cmd, args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunDiff(out, errOut, fs)
		},
	}

	cmd.Flags().StringVarP(&o.manifestPath, "filename", "f", "", "Pass in a kustomize.yaml file or a directory that contains the file.")
	cmd.MarkFlagRequired("filename")
	return cmd
}

// Validate validates diff command.
func (o *diffOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.New("The diff command takes no arguments.")
	}
	return nil
}

// Complete completes diff command.
func (o *diffOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunDiff gets the differences between Application.Resources() and Application.RawResources().
func (o *diffOptions) RunDiff(out, errOut io.Writer, fs fs.FileSystem) error {
	printer := util.Printer{}
	diff := util.DiffProgram{
		Exec:   exec.New(),
		Stdout: out,
		Stderr: errOut,
	}

	l := loader.Init([]loader.SchemeLoader{loader.NewFileLoader(fs)})

	absPath, err := filepath.Abs(o.manifestPath)
	if err != nil {
		return err
	}

	rootLoader, err := l.New(absPath)
	if err != nil {
		return err
	}

	application, err := app.New(rootLoader)
	if err != nil {
		return err
	}
	resources, err := application.Resources()
	if err != nil {
		return err
	}
	rawResources, err := application.RawResources()
	if err != nil {
		return err
	}

	transformedDir, err := util.WriteToDir(resources, "transformed", printer)
	if err != nil {
		return err
	}
	defer transformedDir.Delete()

	noopDir, err := util.WriteToDir(rawResources, "noop", printer)
	if err != nil {
		return err
	}
	defer noopDir.Delete()

	return diff.Run(noopDir.Name, transformedDir.Name)
}
