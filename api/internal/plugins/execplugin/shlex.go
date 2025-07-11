// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin

import (
	"fmt"

	shlex "github.com/carapace-sh/carapace-shlex"
)

func ShlexSplit(s string) ([]string, error) {
	// return shlexSplit(s)
	tokens, err := shlex.Split(s)
	if err != nil {
		return nil, fmt.Errorf("shlex split error: %w", err)
	}
	return tokens.Strings(), nil
}
