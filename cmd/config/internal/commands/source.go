// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GetSourceRunner returns a command for Source.
func GetSourceRunner(name string) *SourceRunner {
	r := &SourceRunner{}
	c := &cobra.Command{
		Use:     "source DIR",
		Short:   commands.SourceShort,
		Long:    commands.SourceLong,
		Example: commands.SourceExamples,
		RunE:    r.runE,
	}
	fixDocs(name, c)
	c.Flags().StringVar(&r.WrapKind, "wrap-kind", kio.ResourceListKind,
		"output using this format.")
	c.Flags().StringVar(&r.WrapApiVersion, "wrap-version", kio.ResourceListAPIVersion,
		"output using this format.")
	c.Flags().StringVar(&r.FunctionConfig, "function-config", "",
		"path to function config.")
	r.Command = c
	_ = c.MarkFlagFilename("function-config", "yaml", "json", "yml")
	return r
}

func SourceCommand(name string) *cobra.Command {
	return GetSourceRunner(name).Command
}

// SourceRunner contains the run function
type SourceRunner struct {
	WrapKind       string
	WrapApiVersion string
	FunctionConfig string
	Command        *cobra.Command
}

func (r *SourceRunner) runE(c *cobra.Command, args []string) error {
	// if there is a function-config specified, emit it
	var functionConfig *yaml.RNode
	if r.FunctionConfig != "" {
		configs, err := kio.LocalPackageReader{PackagePath: r.FunctionConfig}.Read()
		if err != nil {
			return err
		}
		if len(configs) != 1 {
			return fmt.Errorf("expected exactly 1 functionConfig, found %d", len(configs))
		}
		functionConfig = configs[0]
	}

	var outputs []kio.Writer
	outputs = append(outputs, kio.ByteWriter{
		Writer:                c.OutOrStdout(),
		KeepReaderAnnotations: true,
		WrappingKind:          r.WrapKind,
		WrappingAPIVersion:    r.WrapApiVersion,
		FunctionConfig:        functionConfig,
	})

	var inputs []kio.Reader
	for _, a := range args {
		inputs = append(inputs, kio.LocalPackageReader{PackagePath: a})
	}
	if len(inputs) == 0 {
		inputs = []kio.Reader{&kio.ByteReader{Reader: c.InOrStdin()}}
	}

	err := kio.Pipeline{Inputs: inputs, Outputs: outputs}.Execute()
	return handleError(c, err)
}
