// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate $GOBIN/mdtogo docs/commands generateddocs/commands --license=none

package status

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/resource/status/cmd"
)

func StatusCommand() *cobra.Command {
	var status = &cobra.Command{
		Use:   "status",
		Short: "[Alpha] Commands for working with resource status.",
	}

	status.AddCommand(cmd.FetchCommand())
	status.AddCommand(cmd.WaitCommand())
	status.AddCommand(cmd.EventsCommand())

	return status
}
