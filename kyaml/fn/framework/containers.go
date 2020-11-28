// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"bytes"
	"text/template"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

// PatchContainersWithString executes t as a template and patches each container in each resource
// with the result.
func PatchContainersWithString(resources []*yaml.RNode, t string, input interface{}, containers ...string) error {
	resourcePatch := template.Must(template.New("containers").Parse(t))
	return PatchContainersWithTemplate(resources, resourcePatch, input, containers...)
}

// PatchContainersWithTemplate executes t and patches each container in each resource
// with the result.
func PatchContainersWithTemplate(resources []*yaml.RNode, t *template.Template, input interface{}, containers ...string) error {
	var b bytes.Buffer
	if err := t.Execute(&b, input); err != nil {
		return errors.Wrap(err)
	}
	patch, err := yaml.Parse(b.String())
	if err != nil {
		return errors.WrapPrefixf(err, b.String())
	}
	return PatchContainers(resources, patch, containers...)
}

// PatchContainers applies patch to each container in each resource.
func PatchContainers(resources []*yaml.RNode, patch *yaml.RNode, containers ...string) error {
	names := sets.String{}
	names.Insert(containers...)

	for i := range resources {
		containers, err := resources[i].Pipe(yaml.Lookup("spec", "template", "spec", "containers"))
		if err != nil {
			return errors.Wrap(err)
		}
		if containers == nil {
			continue
		}
		err = containers.VisitElements(func(node *yaml.RNode) error {
			f := node.Field("name")
			if f == nil {
				return nil
			}
			if names.Len() > 0 && !names.Has(yaml.GetValue(f.Value)) {
				return nil
			}
			_, err := merge2.Merge(patch, node, yaml.MergeOptions{})
			return errors.Wrap(err)
		})
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}
