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
	"strings"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/validate"
	"github.com/spf13/cobra"
)

// KindOfAdd is the kind of metadata being added: label or annotation
type KindOfAdd int

const (
	annotation KindOfAdd = iota
	label
)

func (k KindOfAdd) String() string {
	kinds := [...]string{
		"annotation",
		"label",
	}
	if k < 0 || k > 1 {
		return "Unknown metadatakind"
	}
	return kinds[k]
}

type addMetadataOptions struct {
	metadata map[string]string
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
			err := o.ValidateAndParse(args, annotation)
			if err != nil {
				return err
			}
			return o.RunAddAnnotation(fsys, annotation)
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
			err := o.ValidateAndParse(args, label)
			if err != nil {
				return err
			}
			return o.RunAddLabel(fsys, label)
		},
	}
	return cmd
}

// ValidateAndParse validates addLabel and addAnnotation commands and parses them into o.metadata
func (o *addMetadataOptions) ValidateAndParse(args []string, k KindOfAdd) error {
	o.metadata = make(map[string]string)
	if len(args) < 1 {
		return fmt.Errorf("must specify %s", k)
	}
	if len(args) > 1 {
		return fmt.Errorf("%ss must be comma-separated, with no spaces. See help text for example", k)
	}
	inputs := strings.Split(args[0], ",")
	for _, input := range inputs {
		switch k {
		case label:
			valid, err := validate.IsValidLabel(input)
			if !valid {
				return err
			}
		case annotation:
			valid, err := validate.IsValidAnnotation(input)
			if !valid {
				return err
			}
		default:
			return fmt.Errorf("unknown metadata kind %s", k)
		}
		//parse annotation keys and values into metadata
		kv := strings.Split(input, ":")
		o.metadata[kv[0]] = kv[1]
	}
	return nil
}

// RunAddAnnotation runs addAnnotation command, doing the real work.
func (o *addMetadataOptions) RunAddAnnotation(fsys fs.FileSystem, k KindOfAdd) error {
	mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
	if err != nil {
		return err
	}
	m, err := mf.read()
	if err != nil {
		return err
	}

	if m.CommonAnnotations == nil {
		m.CommonAnnotations = make(map[string]string)
	}

	for key, value := range o.metadata {
		if k == annotation {
			if _, ok := m.CommonAnnotations[key]; ok {
				return fmt.Errorf("%s %s already in kustomization file", k, key)
			}
			m.CommonAnnotations[key] = value
		}
	}
	return mf.write(m)
}

// RunAddLabel runs addLabel command, doing the real work.
func (o *addMetadataOptions) RunAddLabel(fsys fs.FileSystem, k KindOfAdd) error {
	mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
	if err != nil {
		return err
	}
	m, err := mf.read()
	if err != nil {
		return err
	}

	if m.CommonLabels == nil {
		m.CommonLabels = make(map[string]string)
	}

	for key, value := range o.metadata {
		if _, ok := m.CommonLabels[key]; ok {
			return fmt.Errorf("%s %s already in kustomization file", k, key)
		}
		m.CommonLabels[key] = value
	}
	return mf.write(m)
}
