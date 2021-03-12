// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

func AddFlagEnableTransformerMode(set *pflag.FlagSet) {
	set.BoolVar(
		&theFlags.enable.transformerMode,
		"as-transformer", // flag name
		false,            // default value
		"enable ability to use yamls from stdin as resources.", // help
	)
}
