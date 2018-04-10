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
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"k8s.io/kubectl/pkg/kustomize/constants"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

type addResourceOptions struct {
	resourceFilePath string
}

// newCmdAddResource adds the name of a file containing a resource to the manifest.
func newCmdAddResource(out, errOut io.Writer, fsys fs.FileSystem) *cobra.Command {
	var o addResourceOptions

	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Add the name of a file containing a resource to the manifest.",
		Long:  "Add the name of a file containing a resource to the manifest.",
		Example: `
		add resource {filepath}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunAddResource(out, errOut, fsys)
		},
	}
	return cmd
}

// Validate validates addResource command.
func (o *addResourceOptions) Validate(args []string) error {
	if len(args) != 1 {
		return errors.New("must specify a resource file")
	}
	o.resourceFilePath = args[0]
	return nil
}

// Complete completes addResource command.
func (o *addResourceOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// RunAddResource runs addResource command (do real work).
func (o *addResourceOptions) RunAddResource(out, errOut io.Writer, fsys fs.FileSystem) error {
	_, err := fsys.Stat(o.resourceFilePath)
	if err != nil {
		return err
	}

	mf, err := newManifestFile(constants.KustomizeFileName, fsys)
	if err != nil {
		return err
	}

	m, err := mf.read()
	if err != nil {
		return err
	}

	if stringInSlice(o.resourceFilePath, m.Resources) {
		return fmt.Errorf("resource %s already in manifest", o.resourceFilePath)
	}

	m.Resources = append(m.Resources, o.resourceFilePath)

	return mf.write(m)
}
