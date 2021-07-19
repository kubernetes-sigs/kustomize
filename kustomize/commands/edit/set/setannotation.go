// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setAnnotationOptions struct {
	metadata     map[string]string
	mapValidator func(map[string]string) error
}

// IsValidKey checks key against regex. First part for prefix segment (DNS1123Label) of an annotation followed by a slash, second part for name segment of an annotation
// see https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
var IsValidKey = regexp.MustCompile(`^([a-zA-Z](([-a-zA-Z0-9.]{0,251})[a-zA-Z0-9])?\/)?[a-zA-Z0-9]([-a-zA-Z0-9_.]{0,61}[a-zA-Z0-9])?$`).MatchString

// newCmdSetAnnotation sets one or more commonAnnotations to the kustomization file.
func newCmdSetAnnotation(fSys filesys.FileSystem, v func(map[string]string) error) *cobra.Command {
	var o setAnnotationOptions
	o.mapValidator = v
	cmd := &cobra.Command{
		Use: "annotation",
		Short: "Sets one or more commonAnnotations in " +
			konfig.DefaultKustomizationFileName(),
		Example: `
		set annotation {annotationKey1:annotationValue1} {annotationKey2:annotationValue2}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runE(args, fSys, o.setAnnotations)
		},
	}
	return cmd
}

func (o *setAnnotationOptions) runE(
	args []string, fSys filesys.FileSystem, setter func(*types.Kustomization) error) error {
	err := o.validateAndParse(args)
	if err != nil {
		return err
	}
	kf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := kf.Read()
	if err != nil {
		return err
	}
	err = setter(m)
	if err != nil {
		return err
	}
	return kf.Write(m)
}

// validateAndParse validates `set` commands and parses them into o.metadata
func (o *setAnnotationOptions) validateAndParse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specify annotation")
	}
	m, err := util.ConvertSliceToMap(args, "annotation")
	if err != nil {
		return err
	}
	if err = o.mapValidator(m); err != nil {
		return err
	}
	for key := range m {
		if !IsValidKey(key) {
			return errors.New("invalid annotation key: see the syntax and character set rules at https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/")
		}
	}
	o.metadata = m
	return nil
}

func (o *setAnnotationOptions) setAnnotations(m *types.Kustomization) error {
	if m.CommonAnnotations == nil {
		m.CommonAnnotations = make(map[string]string)
	}
	return o.writeToMap(m.CommonAnnotations)
}

func (o *setAnnotationOptions) writeToMap(m map[string]string) error {
	for k, v := range o.metadata {
		m[k] = v
	}
	return nil
}
