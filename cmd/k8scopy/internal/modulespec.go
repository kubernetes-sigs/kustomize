// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"io/ioutil"

	"sigs.k8s.io/yaml"
)

type ModuleSpec struct {
	Module   string        `json:"module,omitempty" yaml:"module,omitempty"`
	Version  string        `json:"version,omitempty" yaml:"version,omitempty"`
	Packages []PackageSpec `json:"packages,omitempty" yaml:"packages,omitempty"`
}

func (s ModuleSpec) Name() string {
	return s.Module + "@" + s.Version
}

type PackageSpec struct {
	Name  string   `json:"name,omitempty" yaml:"name,omitempty"`
	Files []string `json:"files,omitempty" yaml:"files,omitempty"`
}

func ReadSpec(fileName string) *ModuleSpec {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	var spec ModuleSpec
	if err = yaml.Unmarshal(bytes, &spec); err != nil {
		panic(err)
	}
	return &spec
}
