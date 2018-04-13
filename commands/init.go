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

	"errors"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/kustomize/constants"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

const kustomizationTemplate = `
namePrefix: some-prefix
# Labels to add to all objects and selectors.
# These labels would also be used to form the selector for apply --prune
# Named differently than “labels” to avoid confusion with metadata for this object
labelsToAdd:
  app: helloworld
annotationsToAdd:
  note: This is an example annotation
resources: []
#- service.yaml
#- ../some-dir/
# There could also be configmaps in Base, which would make these overlays
configMapGenerator: []
# There could be secrets in Base, if just using a fork/rebase workflow
secretGenerator: []
`

type initOptions struct {
}

// NewCmdInit makes the init command.
func newCmdInit(out, errOut io.Writer, fs fs.FileSystem) *cobra.Command {
	var o initOptions

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Creates a file called \"" + constants.KustomizationFileName + "\" in the current directory",
		Long: "Creates a file called \"" +
			constants.KustomizationFileName + "\" in the current directory with example values.",
		Example:      `init`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(cmd, args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunInit(out, errOut, fs)
		},
	}
	return cmd
}

// Validate validates init command.
func (o *initOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.New("The init command takes no arguments.")
	}
	return nil
}

// Complete completes init command.
func (o *initOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunInit writes a kustomization file.
func (o *initOptions) RunInit(out, errOut io.Writer, fs fs.FileSystem) error {
	if _, err := fs.Stat(constants.KustomizationFileName); err == nil {
		return fmt.Errorf("%q already exists", constants.KustomizationFileName)
	}
	return fs.WriteFile(constants.KustomizationFileName, []byte(kustomizationTemplate))
}
