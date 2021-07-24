// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"strconv"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GenArgs is a facade over GeneratorArgs, exposing a few readonly properties.
type GenArgs struct {
	args *GeneratorArgs
}

// NewGenArgs returns a new instance of GenArgs.
func NewGenArgs(args *GeneratorArgs) *GenArgs {
	return &GenArgs{args: args}
}

func (g *GenArgs) String() string {
	if g == nil {
		return "{nilGenArgs}"
	}
	return "{" +
		strings.Join([]string{
			"nsfx:" + strconv.FormatBool(g.ShouldAddHashSuffixToName()),
			"beh:" + g.Behavior().String()},
			",") +
		"}"
}

// ShouldAddHashSuffixToName returns true if a resource
// content hash should be appended to the name of the resource.
func (g *GenArgs) ShouldAddHashSuffixToName() bool {
	return g.args != nil &&
		(g.args.Options == nil || !g.args.Options.DisableNameSuffixHash)
}

// Behavior returns Behavior field of GeneratorArgs
func (g *GenArgs) Behavior() GenerationBehavior {
	if g == nil || g.args == nil {
		return BehaviorUnspecified
	}
	return NewGenerationBehavior(g.args.Behavior)
}

// IsNilOrEmpty returns true if g is nil or if the args are empty
func (g *GenArgs) IsNilOrEmpty() bool {
	return g == nil || g.args == nil
}

// AsYaml returns a yaml marshalling of the underlying Genargs
func (g *GenArgs) AsYaml() ([]byte, error) {
	if g == nil {
		return yaml.Marshal(nil)
	}
	return yaml.Marshal(g.args)
}
