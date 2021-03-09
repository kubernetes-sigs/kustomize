// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

func AddFlagEnableGitCache(set *pflag.FlagSet) {
	set.BoolVar(
		&theFlags.enable.gitCache,
		"enable-git-cache",
		false,  // default
		"Enable caching cloned git repositories. (experiment)")
}
