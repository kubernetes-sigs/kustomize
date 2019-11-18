// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"strconv"
	"strings"
)

// GenArgs contains both GeneratorArgs and GeneratorOptions.
type GenArgs struct {
	args *GeneratorArgs
	opts *GeneratorOptions
}

// NewGenArgs returns a new object of GenArgs
func NewGenArgs(args *GeneratorArgs, opts *GeneratorOptions) *GenArgs {
	return &GenArgs{
		args: args,
		opts: opts,
	}
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
		(g.opts == nil || !g.opts.DisableNameSuffixHash)
}

// Behavior returns Behavior field of GeneratorArgs
func (g *GenArgs) Behavior() GenerationBehavior {
	if g.args == nil {
		return BehaviorUnspecified
	}
	return NewGenerationBehavior(g.args.Behavior)
}
