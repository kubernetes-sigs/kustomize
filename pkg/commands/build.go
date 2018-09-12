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

	"github.com/spf13/cobra"

	"errors"

	"sigs.k8s.io/kustomize/pkg/app"
	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
)

type buildOptions struct {
	kustomizationPath string
	outputPath        string
}

var examples = `
Use the file somedir/kustomization.yaml to generate a set of api resources:
    build somedir

Use a url pointing to a remote directory/kustomization.yaml to generate a set of api resources:
    build url
The url should follow hashicorp/go-getter URL format described in
https://github.com/hashicorp/go-getter#url-format

url examples:
  sigs.k8s.io/kustomize//examples/multibases?ref=v1.0.6
  github.com/Liujingfang1/mysql
  github.com/Liujingfang1/kustomize//examples/helloWorld?ref=repoUrl2
`

// newCmdBuild creates a new build command.
func newCmdBuild(out io.Writer, fs fs.FileSystem) *cobra.Command {
	var o buildOptions

	cmd := &cobra.Command{
		Use:          "build [path]",
		Short:        "Print current configuration per contents of " + constants.KustomizationFileName,
		Example:      examples,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunBuild(out, fs)
		},
	}
	cmd.Flags().StringVarP(
		&o.outputPath,
		"output", "o", "",
		"If specified, write the build output to this path.")
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
func (o *buildOptions) RunBuild(out io.Writer, fSys fs.FileSystem) error {
	rootLoader, err := loader.NewLoader(o.kustomizationPath, "", fSys)
	if err != nil {
		return err
	}
	defer rootLoader.Cleanup()

	application, err := app.NewApplication(rootLoader, fSys)
	if err != nil {
		return err
	}

	allResources, err := application.MakeCustomizedResMap()

	if err != nil {
		return err
	}

	// Output the objects.
	res, err := allResources.EncodeAsYaml()
	if err != nil {
		return err
	}

	if o.outputPath != "" {
		return fSys.WriteFile(o.outputPath, res)
	}
	_, err = out.Write(res)
	return err
}
