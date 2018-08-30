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
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
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
	metadata            map[string]string
	validateAnnotations func(map[string]string) field.ErrorList
	validateLabels      func(map[string]string) field.ErrorList
}

// newCmdAddAnnotation adds one or more commonAnnotations to the kustomization file.
func newCmdAddAnnotation(fsys fs.FileSystem) *cobra.Command {
	var o addMetadataOptions
	o.validateAnnotations = func(x map[string]string) field.ErrorList {
		return validation.ValidateAnnotations(x, field.NewPath("field"))
	}

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
	o.validateLabels = func(x map[string]string) field.ErrorList {
		return v1validation.ValidateLabels(x, field.NewPath("field"))
	}

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
		metadata := make(map[string]string)
		//parse annotation keys and values into metadata
		kv := strings.Split(input, ":")
		if len(kv[0]) < 1 {
			return fmt.Errorf("invalid %s format: %s", k, input)
		}
		if len(kv) > 1 {
			metadata[kv[0]] = kv[1]
		} else {
			metadata[kv[0]] = ""
		}
		switch k {
		case label:
			if errs := o.validateLabels(metadata); len(errs) != 0 {
				return fmt.Errorf("invalid %s format: %s", k, input)
			}
		case annotation:
			if errs := o.validateAnnotations(metadata); len(errs) != 0 {
				return fmt.Errorf("invalid %s format: %s", k, input)
			}
		default:
			return fmt.Errorf("unknown metadata kind %s", k)
		}
		o.metadata[kv[0]] = metadata[kv[0]]
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
