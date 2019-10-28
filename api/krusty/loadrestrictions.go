// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty

import (
	"fmt"
	"github.com/spf13/pflag"
)

//go:generate stringer -type=loadRestrictions
type loadRestrictions int

const (
	unknown loadRestrictions = iota

	// With this restriction, the files referenced by a
	// kustomization file must be in or under the directory
	// holding the kustomization file itself.
	rootOnly

	// The kustomization file may specify absolute or
	// relative paths to patch or resources files outside
	// its own tree.
	none
)

const (
	flagName = "load_restrictor"
)

var (
	flagValue = rootOnly.String()
	flagHelp  = "if set to '" + none.String() +
		"', local kustomizations may load files from outside their root. " +
		"This does, however, break the relocatability of the kustomization."
)

func AddFlagLoadRestrictor(set *pflag.FlagSet) {
	set.StringVar(
		&flagValue, flagName,
		rootOnly.String(), flagHelp)
}

func ValidateFlagLoadRestrictor() error {
	switch flagValue {
	case rootOnly.String():
		return nil
	case none.String():
		return nil
	default:
		return fmt.Errorf(
			"illegal flag value --%s %s; legal values: %v",
			flagName, flagValue,
			[]string{rootOnly.String(), none.String()})
	}
}

func GetFlagLoadRestrictorValue() loadRestrictions {
	switch flagValue {
	case rootOnly.String():
		return rootOnly
	case none.String():
		return none
	default:
		return unknown
	}
}
