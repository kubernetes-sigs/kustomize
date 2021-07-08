// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/krmfile"
)

// GetInitRunner returns a command InitRunner.
func GetInitRunner(name string) *InitRunner {
	r := &InitRunner{}
	c := &cobra.Command{
		Use:     "init DIR...",
		Args:    cobra.RangeArgs(0, 1),
		Short:   commands.InitShort,
		Long:    commands.InitLong,
		Example: commands.InitExamples,
		RunE:    r.runE,
		Deprecated: "setter commands and substitutions will no longer be available in kustomize v5.\n" +
			"See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.",
	}
	runner.FixDocs(name, c)
	r.Command = c
	return r
}

func InitCommand(name string) *cobra.Command {
	return GetInitRunner(name).Command
}

// InitRunner contains the init function
type InitRunner struct {
	Command *cobra.Command
}

func (r *InitRunner) runE(c *cobra.Command, args []string) error {
	var dir string
	if len(args) == 0 {
		dir = "."
	} else {
		dir = args[0]
	}
	filename := filepath.Join(dir, krmfile.KrmfileName)

	if _, err := os.Stat(filename); err == nil || !os.IsNotExist(err) {
		return errors.Errorf("directory already initialized with a Krmfile")
	}

	return ioutil.WriteFile(filename, []byte(strings.TrimSpace(`
apiVersion: config.k8s.io/v1alpha1
kind: Krmfile
`)), 0600)
}
