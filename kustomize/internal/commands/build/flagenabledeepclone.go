// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

const (
	flagEnableDeepCloneName = "enable_deep_clone"
	flagEnableDeepCloneHelp = `enable deep git cloning.
See https://github.com/kubernetes-sigs/kustomize/issues/1452
`
)

var (
	flagEnableDeepCloneValue = false
)

func addFlagEnableDeepClone(set *pflag.FlagSet) {
	set.BoolVar(
		&flagEnableDeepCloneValue, flagEnableDeepCloneName,
		false, flagEnableDeepCloneHelp)
}

func isFlagEnableDeepCloneSet() bool {
	return flagEnableDeepCloneValue
}
