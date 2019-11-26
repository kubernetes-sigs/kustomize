// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate $GOBIN/mdtogo docs/api-conventions cmddocs/api --full=true --license=none
//go:generate $GOBIN/mdtogo docs/tutorials cmddocs/tutorials --full=true --license=none
//go:generate $GOBIN/mdtogo docs/commands cmddocs/commands --license=none
package main

import (
	"os"

	"sigs.k8s.io/kustomize/cmd/config/cmds"
)

func main() {
	if err := cmds.NewConfigCommand("").Execute(); err != nil {
		os.Exit(1)
	}
}
