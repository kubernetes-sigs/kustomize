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
	"github.com/kubernetes-sigs/kustomize/pkg/validators"
	"github.com/spf13/cobra"
)

// kindOfAdd is the kind of metadata being added: label or annotation
type kindOfAdd int

const (
	annotation kindOfAdd = iota
	label
)

func (k kindOfAdd) String() string {
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
	metadata     map[string]string
	mapValidator validators.MapValidatorFunc
	kind         kindOfAdd
}

// newCmdAddAnnotation adds one or more commonAnnotations to the kustomization file.
func newCmdAddAnnotation(fSys fs.FileSystem, v validators.MapValidatorFunc) *cobra.Command {
	var o addMetadataOptions
	o.kind = annotation
	o.mapValidator = v

	cmd := &cobra.Command{
		Use:   "annotation",
		Short: "Adds one or more commonAnnotations to " + constants.KustomizationFileName,
		Example: `
		add annotation {annotationKey1:annotationValue1},{annotationKey2:annotationValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.ValidateAndParse(args)
			if err != nil {
				return err
			}
			return o.RunAddAnnotation(fSys)
		},
	}
	return cmd
}

// newCmdAddLabel adds one or more commonLabels to the kustomization file.
func newCmdAddLabel(fSys fs.FileSystem, v validators.MapValidatorFunc) *cobra.Command {
	var o addMetadataOptions
	o.kind = label
	o.mapValidator = v

	cmd := &cobra.Command{
		Use:   "label",
		Short: "Adds one or more commonLabels to " + constants.KustomizationFileName,
		Example: `
		add label {labelKey1:labelValue1},{labelKey2:labelValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.ValidateAndParse(args)
			if err != nil {
				return err
			}
			return o.RunAddLabel(fSys)
		},
	}
	return cmd
}

// ValidateAndParse validates addLabel and addAnnotation commands and parses them into o.metadata
func (o *addMetadataOptions) ValidateAndParse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specify %s", o.kind)
	}
	if len(args) > 1 {
		return fmt.Errorf("%ss must be comma-separated, with no spaces", o.kind)
	}
	m, err := o.convertToMap(args[0])
	if err != nil {
		return err
	}
	if err = o.mapValidator(m); err != nil {
		return err
	}
	o.metadata = m
	return nil
}

func (o *addMetadataOptions) convertToMap(arg string) (map[string]string, error) {
	result := make(map[string]string)
	inputs := strings.Split(arg, ",")
	for _, input := range inputs {
		kv := strings.Split(input, ":")
		if len(kv[0]) < 1 {
			return nil, makeError(o.kind, input, "empty key")
		}
		if len(kv) > 2 {
			return nil, makeError(o.kind, input, "too many colons")
		}
		if len(kv) > 1 {
			result[kv[0]] = kv[1]
		} else {
			result[kv[0]] = ""
		}
	}
	return result, nil
}

// RunAddAnnotation runs addAnnotation command, doing the real work.
func (o *addMetadataOptions) RunAddAnnotation(fsys fs.FileSystem) error {
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
		if _, ok := m.CommonAnnotations[key]; ok {
			return fmt.Errorf("%s %s already in kustomization file", o.kind, key)
		}
		m.CommonAnnotations[key] = value
	}
	return mf.write(m)
}

// RunAddLabel runs addLabel command, doing the real work.
func (o *addMetadataOptions) RunAddLabel(fsys fs.FileSystem) error {
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
			return fmt.Errorf("%s %s already in kustomization file", o.kind, key)
		}
		m.CommonLabels[key] = value
	}
	return mf.write(m)
}

func makeError(k kindOfAdd, input string, message string) error {
	return fmt.Errorf("invalid %s: %s (%s)", k, input, message)
}
