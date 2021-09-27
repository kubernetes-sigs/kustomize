// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
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
		PreRunE: r.preRunE,
		Args:    cobra.MaximumNArgs(1),
	}
	runner.FixDocs(name, c)
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

func (r *SinkRunner) preRunE(c *cobra.Command, args []string) error {
	_, err := fmt.Fprintln(os.Stderr, `Command "sink" is deprecated, this will no longer be available in kustomize v5.
See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.`)
	return err
}

func (r *SinkRunner) runE(c *cobra.Command, args []string) error {
	var outputs []kio.Writer
	if len(args) == 1 {
		outputs = []kio.Writer{&kio.LocalPackageWriter{PackagePath: args[0]}}
	} else {
		outputs = []kio.Writer{&kio.ByteWriter{
			Writer:           c.OutOrStdout(),
			ClearAnnotations: []string{kioutil.PathAnnotation, kioutil.LegacyPathAnnotation}},
		}
	}

	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: c.InOrStdin()}},
		Outputs: outputs}.Execute()
	return runner.HandleError(c, err)
}
