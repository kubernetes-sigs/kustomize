// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commandutil

import (
	"os"
)

// EnabkeAlphaCommmandsEnvName is the environment variable used to enable Alpha kustomize commands.
//If set to "true" alpha commands will be enabled.
const EnableAlphaCommmandsEnvName = "KUSTOMIZE_ENABLE_ALPHA_COMMANDS"

// GetAlphaEnabled returns true if alpha commands should be enabled.
func GetAlphaEnabled() bool {
	return os.Getenv(EnableAlphaCommmandsEnvName) == "true"
}
