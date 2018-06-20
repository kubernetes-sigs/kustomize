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
	"regexp"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

type addLabelOptions struct {
	labels string
}

// newCmdAddLabel adds one or more commonLabels to the kustomization file.
func newCmdAddLabel(fsys fs.FileSystem) *cobra.Command {
	var o addLabelOptions

	cmd := &cobra.Command{
		Use:   "label",
		Short: "Adds one or more commonLabels to the kustomization.yaml in current directory",
		Example: `
		add label {labelKey1:labelValue1},{labelKey2:labelValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunAddLabel(fsys)
		},
	}
	return cmd
}

// Validate validates addLabel command.
// TODO: make sure label is of correct format key:value
func (o *addLabelOptions) Validate(args []string) error {
	for _, arg := range args {
		fmt.Println(arg + "****")
	}

	if len(args) < 1 {
		return errors.New("must specify a label")
	}
	if len(args) > 1 {
		return errors.New("labels must be comma-separated, with no spaces. See help text for example.")
	}
	inputs := strings.Split(args[0],",")
	for _, input := range inputs {
			ok, err := regexp.MatchString(`\A([a-zA-Z0-9_.-]+):([a-zA-Z0-9_.-]+)\z`, input)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("invalid label format: %s", input)
		}

	}

	o.labels = args[0]
	return nil
}

// Complete completes addLabel command.
func (o *addLabelOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunAddLabel runs addLabel command (do real work).
func (o *addLabelOptions) RunAddLabel(fsys fs.FileSystem) error {
	mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
	if err != nil {
		return err
	}

	m, err := mf.read()
	if err != nil {
		return err
	}

	if m.CommonLabels == nil{
		m.CommonLabels = make(map[string]string)
	}

	labels := strings.Split(o.labels, ",")
	for _, label := range labels {
		kv := strings.Split(label, ":")
		if _, ok := m.CommonLabels[kv[0]]; ok {
			return fmt.Errorf("label %s already in kustomization file", kv[0])
		}
		m.CommonLabels[kv[0]] = kv[1]
	}

	return mf.write(m)
}
