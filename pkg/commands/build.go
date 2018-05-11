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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"errors"

	"k8s.io/kubectl/pkg/kustomize/app"
	"k8s.io/kubectl/pkg/kustomize/constants"
	kutil "k8s.io/kubectl/pkg/kustomize/util"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
	"k8s.io/kubectl/pkg/loader"
)

type buildOptions struct {
	kustomizationPath string
}

// newCmdBuild creates a new build command.
func newCmdBuild(out, errOut io.Writer, fs fs.FileSystem) *cobra.Command {
	var o buildOptions

	cmd := &cobra.Command{
		Use:   "build [path]",
		Short: "Print current configuration per contents of " + constants.KustomizationFileName,
		Example: "Use the file somedir/" + constants.KustomizationFileName +
			" to generate a set of api resources:\nbuild somedir/",
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Validate(args)
			if err != nil {
				fmt.Fprintf(errOut, "error: %v\n", err)
				os.Exit(1)
			}
			err = o.RunBuild(out, errOut, fs)
			if err != nil {
				fmt.Fprintf(errOut, "error: %v\n", err)
				os.Exit(1)
			}
		},
	}
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
func (o *buildOptions) RunBuild(out, errOut io.Writer, fs fs.FileSystem) error {
	l := loader.Init([]loader.SchemeLoader{loader.NewFileLoader(fs)})

	absPath, err := filepath.Abs(o.kustomizationPath)
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

	allResources, err := application.Resources()

	if err != nil {
		return err
	}

	// Output the objects.
	res, err := kutil.Encode(allResources)
	if err != nil {
		return err
	}
	_, err = out.Write(res)
	return err
}
