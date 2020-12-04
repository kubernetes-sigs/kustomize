// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

const (
	flagAllowResourceIdChangesName = "allow_id_changes"
	flagAllowResourceIdChangesHelp = `enable changes to a resourceId`
)

var (
	flagAllowResourceIdChangesValue = false
)

func addFlagAllowResourceIdChanges(set *pflag.FlagSet) {
	set.BoolVar(
		&flagAllowResourceIdChangesValue, flagAllowResourceIdChangesName,
		false, flagAllowResourceIdChangesHelp)
}
