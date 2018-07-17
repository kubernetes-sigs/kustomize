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
	"fmt"
	"regexp"
	"strings"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/spf13/cobra"
)

const label = "label"
const ann = "annotation"

type addMetadataOptions struct {
	metadata string
}

// newCmdAddAnnotation adds one or more commonAnnotations to the kustomization file.
func newCmdAddAnnotation(fsys fs.FileSystem) *cobra.Command {
	var o addMetadataOptions

	cmd := &cobra.Command{
		Use:   "annotation",
		Short: "Adds one or more commonAnnotations to the kustomization.yaml in current directory",
		Example: `
		add annotation {annotationKey1:annotationValue1},{annotationKey2:annotationValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mdKind := ann
			err := o.Validate(args, mdKind)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunAddMetadata(fsys, mdKind)
		},
	}
	return cmd
}

// newCmdAddLabel adds one or more commonLabels to the kustomization file.
func newCmdAddLabel(fsys fs.FileSystem) *cobra.Command {
	var o addMetadataOptions

	cmd := &cobra.Command{
		Use:   "label",
		Short: "Adds one or more commonLabels to the kustomization.yaml in current directory",
		Example: `
		add label {labelKey1:labelValue1},{labelKey2:labelValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mdKind := label
			err := o.Validate(args, mdKind)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunAddMetadata(fsys, mdKind)
		},
	}
	return cmd
}

// Validate validates addLabel and addAnnotation commands.
func (o *addMetadataOptions) Validate(args []string, mdKind string) error {

	if len(args) < 1 {
		return fmt.Errorf("must specify %s", mdKind)
	}
	if len(args) > 1 {
		return fmt.Errorf("%ss must be comma-separated, with no spaces. See help text for example", mdKind)
	}
	inputs := strings.Split(args[0], ",")
	for _, input := range inputs {
		ok, err := regexp.MatchString(`\A([a-zA-Z0-9_.-]+):([a-zA-Z0-9_.-]+)\z`, input)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("invalid %s format: %s", mdKind, input)
		}
	}

	o.metadata = args[0]
	return nil
}

// Complete completes addMetadata command.
func (o *addMetadataOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunAddMetadata runs addLabel and addAnnotation commands (do real work).
func (o *addMetadataOptions) RunAddMetadata(fsys fs.FileSystem, mdKind string) error {
	mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
	if err != nil {
		return err
	}

	m, err := mf.read()
	if err != nil {
		return err
	}

	if mdKind == label && m.CommonLabels == nil {
		m.CommonLabels = make(map[string]string)
	}

	if mdKind == ann && m.CommonAnnotations == nil {
		m.CommonAnnotations = make(map[string]string)
	}

	entries := strings.Split(o.metadata, ",")
	for _, entry := range entries {
		kv := strings.Split(entry, ":")
		if mdKind == label {
			if _, ok := m.CommonLabels[kv[0]]; ok {
				return fmt.Errorf("%s %s already in kustomization file", mdKind, kv[0])
			}
			m.CommonLabels[kv[0]] = kv[1]
		}
		if mdKind == ann {
			if _, ok := m.CommonAnnotations[kv[0]]; ok {
				return fmt.Errorf("%s %s already in kustomization file", mdKind, kv[0])
			}
			m.CommonAnnotations[kv[0]] = kv[1]
		}
	}

	return mf.write(m)
}
