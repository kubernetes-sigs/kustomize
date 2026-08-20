// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

func AddFlagStrictWarningsFlag(set *pflag.FlagSet) {
	set.BoolVar(
		&theFlags.enable.strictWarnings,
		"strict-warnings",
		false,
		"flag to treat warnings as errors, "+
			"i.e. return a non-zero exit code if warnings are emitted.")
}
