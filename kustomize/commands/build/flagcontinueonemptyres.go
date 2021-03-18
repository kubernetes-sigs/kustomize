// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

func AddFlagContinueOnEmptyResult(set *pflag.FlagSet) {
	set.BoolVar(
		&theFlags.fnOptions.ContinueOnEmptyResult,
		"continue-on-empty-result",
		true,
		"don't stop if function returned emply list - emply list will be provided as input for the next function",
	)
}
