// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

const (
	flagMaxParallelAccumulateName = "max_parallel_accumulate"
	flagMaxParallelAccumulateHelp = `Accumulate resources with n threads in parallel.
`
)

var (
	flagMaxParallelAccumulateValue = 1
)

func addFlagMaxParallelAccumulate(set *pflag.FlagSet) {
	set.IntVar(
		&flagMaxParallelAccumulateValue, flagMaxParallelAccumulateName,
		1, flagMaxParallelAccumulateHelp)
}

func getMaxParallelAccumulateValue() int {
	return flagMaxParallelAccumulateValue
}
