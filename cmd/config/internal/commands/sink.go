// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// GetSinkRunner returns a command for Sink.
func GetSinkRunner(name string) *SinkRunner {
	r := &SinkRunner{}
	c := &cobra.Command{
		Use:     "sink DIR",
		Short:   commands.SinkShort,
		Long:    commands.SinkLong,
		Example: commands.SinkExamples,
		RunE:    r.runE,
		Args:    cobra.ExactArgs(1),
	}
	fixDocs(name, c)
	r.Command = c
	return r
}

func SinkCommand(name string) *cobra.Command {
	return GetSinkRunner(name).Command
}

// SinkRunner contains the run function
type SinkRunner struct {
	Command *cobra.Command
}

func (r *SinkRunner) runE(c *cobra.Command, args []string) error {
	err := kio.Pipeline{
		Inputs: []kio.Reader{&kio.ByteReader{Reader: c.InOrStdin()}},
		Outputs: []kio.Writer{
			&kio.LocalPackageWriter{
				PackagePath:      args[0],
				ClearAnnotations: []string{"config.kubernetes.io/path"},
			},
		},
	}.Execute()
	return handleError(c, err)
}
