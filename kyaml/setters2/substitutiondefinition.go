// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"io/ioutil"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type SubstitutionDefinition struct {
	Name    string  `yaml:"name"`
	Pattern string  `yaml:"pattern"`
	Values  []Value `yaml:"value"`
}

type Value struct {
	Marker string `yaml:"marker"`
	Ref    string `yaml:"ref"`
}

func (subd SubstitutionDefinition) AddSubstitutionToFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	y, err := yaml.Parse(string(b))
	if err != nil {
		return err
	}
	if err := y.PipeE(subd); err != nil {
		return err
	}
	out, err := y.String()
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, []byte(out), 0666); err != nil {
		return err
	}
	return nil
}

func (subd SubstitutionDefinition) Filter(object *yaml.RNode) (*yaml.RNode, error) {
	key := DefinitionPrefix + subd.Name

	def, err := object.Pipe(yaml.LookupCreate(
		yaml.MappingNode, "openAPI", "definitions", key, "x-k8s-cli", "substitution"))
	if err != nil {
		return nil, err
	}
	if err := def.PipeE(yaml.FieldSetter{Name: "name", StringValue: subd.Name}); err != nil {
		return nil, err
	}
	if err := def.PipeE(yaml.FieldSetter{Name: "pattern", StringValue: subd.Pattern}); err != nil {
		return nil, err
	}
	return object, nil
}
