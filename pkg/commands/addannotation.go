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
	"errors"
	"fmt"
	"strings"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
)

type addAnnotationOptions struct {
	annotations string
}

// newCmdAddAnnotation adds one or more commonAnnotations to the kustomization file.
func newCmdAddAnnotation(fsys fs.FileSystem) *cobra.Command {
	var o addAnnotationOptions

	cmd := &cobra.Command{
		Use:   "annotation",
		Short: "Adds one or more commonAnnotations to the kustomization.yaml in current directory",
		Example: `
		add annotation {annotationKey1:annotationValue1},{annotationKey2:annotationValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunAddAnnotation(fsys)
		},
	}
	return cmd
}

// Validate validates addAnnotation command.
func (o *addAnnotationOptions) Validate(args []string) error {
	if len(args) < 1 {
		return errors.New("must specify an annotation")
	}
	if len(args) > 1 {
		return errors.New("annotations must be comma-separated, with no spaces. See help text for example.")
	}
	inputs := strings.Split(args[0],",")
	for _, input := range inputs {
			ok, err := regexp.MatchString(`\A([a-zA-Z0-9_.-]+):([a-zA-Z0-9_.-]+)\z`, input)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("invalid annotation format: %s", input)
		}
	}
	o.annotations = args[0]
	return nil
}

// Complete completes addAnnotation command.
func (o *addAnnotationOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// RunAddAnnotation runs addAnnotation command (do real work).
func (o *addAnnotationOptions) RunAddAnnotation(fsys fs.FileSystem) error {
	mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
	if err != nil {
		return err
	}

	m, err := mf.read()
	if err != nil {
		return err
	}

	if m.CommonAnnotations == nil{
		m.CommonAnnotations = make(map[string]string)
	}
	annotations := strings.Split(o.annotations, ",")
	for _, ann := range annotations {
		kv := strings.Split(ann, ":")
		if _, ok := m.CommonAnnotations[kv[0]]; ok {
			return fmt.Errorf("annotation %s already in kustomization file", kv[0])
		}
		m.CommonAnnotations[kv[0]] = kv[1]
	}

	return mf.write(m)
}
