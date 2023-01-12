// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localize

import (
	"fmt"

	"github.com/spf13/pflag"
)

func AddFlagScope(allocatedFlags *flags, set *pflag.FlagSet) {
	// no shorthand to avoid conflation with other flags
	set.StringVar(&allocatedFlags.scope,
		"scope",
		"",
		fmt.Sprintf(`Path to directory inside of which %s is limited to running.
Cannot specify for remote targets, as scope is by default the containing repo.
If not specified for local target, scope defaults to target.
`, cmdName))
}
