// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

func AddFlagDisableStrictParse(set *pflag.FlagSet) {
	set.BoolVar(
		&theFlags.enable.uniqueKeys,
		"enable-unique-keys",
		true,
		`If set enable-unique-keys to true, the resources will be parsed in strict mode which means kubstomize 
will produce an error when there are duplicate keys in resources.`)
}

func isUniqueKeysEnabled() bool {
	return theFlags.enable.uniqueKeys
}
