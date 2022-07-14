// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localize

import (
	"github.com/spf13/pflag"
)

func AddFlagScope(set *pflag.FlagSet) {
	set.StringVarP(
		&theFlags.scope,
		"scope",
		"",
		"", // default
		"If specified, limit target references to inside this path.")
}

func validateFlagScope() error {
	// TODO: think about this, maybe can be empty
	/*if theFlags.scope == "" {
		return errors.Errorf("localize scope flag cannot be empty")
	}*/
	return nil
}
