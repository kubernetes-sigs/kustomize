// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"os"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kustomize/api/konfig"
)

const (
	flagEnableManagedbyLabelName = "enable_managedby_label"
	flagEnableManagedbyLabelHelp = `enable adding ` + konfig.ManagedbyLabelKey
)

var (
	flagEnableManagedbyLabelValue = false
)

func addFlagEnableManagedbyLabel(set *pflag.FlagSet) {
	set.BoolVar(
		&flagEnableManagedbyLabelValue, flagEnableManagedbyLabelName,
		false, flagEnableManagedbyLabelHelp)
}

func isManagedbyLabelEnabled() bool {
	if flagEnableManagedbyLabelValue {
		return true
	}
	enableLabel, isSet := os.LookupEnv(konfig.EnableManagedbyLabelEnv)
	if isSet && enableLabel == "on" {
		return true
	}
	return false
}
