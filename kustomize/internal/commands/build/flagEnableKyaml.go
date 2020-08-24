// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

var (
	flagEnableKyamlValue = false
)

func addFlagEnableKyaml(set *pflag.FlagSet) {
	set.BoolVar(
		&flagEnableKyamlValue,
		"enable_kyaml", // flag name
		false,          // default value
		"enable dependence on kyaml instead of k8sdeps.", // help
	)
}
