// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kustomize/api/konfig"
)

var (
	flagEnableKyamlValue = konfig.FlagEnableKyamlDefaultValue
)

func addFlagEnableKyaml(set *pflag.FlagSet) {
	set.BoolVar(
		&flagEnableKyamlValue,
		"enable_kyaml",                                   // flag name
		konfig.FlagEnableKyamlDefaultValue,               // default value
		"enable dependence on kyaml instead of k8sdeps.", // help
	)
}
