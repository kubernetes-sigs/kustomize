// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

const flagSuppressDeprecationWarningsName = "suppress-deprecation-warnings"

func AddFlagSuppressDeprecationWarnings(set *pflag.FlagSet) {
	set.BoolVar(
		&theFlags.suppressDeprecationWarnings,
		flagSuppressDeprecationWarningsName,
		false,
		"suppress warnings about deprecated fields in kustomization files "+
			"(e.g., bases, commonLabels, vars)")
}
