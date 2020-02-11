// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"io/ioutil"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const DefinitionPrefix = "io.k8s.cli.setters."

type SetterDefinition struct {
	Name  string
	Value string
}

func (sd SetterDefinition) AddSetterToFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	y, err := yaml.Parse(string(b))
	if err != nil {
		return err
	}
	if err := y.PipeE(sd); err != nil {
		return err
	}
	out, err := y.String()
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, []byte(out), 0600); err != nil {
		return err
	}
	return nil
}

func (sd SetterDefinition) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	key := DefinitionPrefix + sd.Name

	def, err := object.Pipe(yaml.LookupCreate(
		yaml.MappingNode, "openAPI", "definitions", key, "x-k8s-cli", "setter"))
	if err != nil {
		return nil, err
	}
	if err := def.PipeE(yaml.FieldSetter{Name: "name", StringValue: sd.Name}); err != nil {
		return nil, err
	}
	if err := def.PipeE(yaml.FieldSetter{Name: "value", StringValue: sd.Value}); err != nil {
		return nil, err
	}
	return object, nil
}
