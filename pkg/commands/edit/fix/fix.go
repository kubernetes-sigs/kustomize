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

package fix

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/pkg/commands/edit/editopts"
	"sigs.k8s.io/kustomize/pkg/commands/kustfile"
	"sigs.k8s.io/kustomize/pkg/fs"
)

type fixOptions struct {
	editopts.Options
}

// NewCmdFix returns an instance of 'fix' subcommand.
func NewCmdFix(fSys fs.FileSystem) *cobra.Command {
	o := fixOptions{}

	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Fix the missing fields in kustomization file",
		Long:  "",
		Example: `
	# Fix the missing and deprecated fields in kustomization file
	kustomize fix

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.ValidateCommon(cmd, args)
			if err != nil {
				return err
			}
			return o.RunFix(fSys)
		},
	}
	return cmd
}

func (o *fixOptions) RunFix(fSys fs.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(o.KustomizationDir, fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	return mf.Write(m)
}
