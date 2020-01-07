// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GetCatRunner returns a command CatRunner.
func GetCatRunner(name string) *CatRunner {
	r := &CatRunner{}
	c := &cobra.Command{
		Use:     "cat DIR...",
		Short:   commands.CatShort,
		Long:    commands.CatLong,
		Example: commands.CatExamples,
		RunE:    r.runE,
	}
	fixDocs(name, c)
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	c.Flags().BoolVar(&r.Format, "format", true,
		"format resource config yaml before printing.")
	c.Flags().BoolVar(&r.KeepAnnotations, "annotate", false,
		"annotate resources with their file origins.")
	c.Flags().StringVar(&r.WrapKind, "wrap-kind", "",
		"if set, wrap the output in this list type kind.")
	c.Flags().StringVar(&r.WrapApiVersion, "wrap-version", "",
		"if set, wrap the output in this list type apiVersion.")
	c.Flags().StringVar(&r.FunctionConfig, "function-config", "",
		"path to function config to put in ResourceList -- only if wrapped in a ResourceList.")
	c.Flags().StringSliceVar(&r.Styles, "style", []string{},
		"yaml styles to apply.  may be 'TaggedStyle', 'DoubleQuotedStyle', 'LiteralStyle', "+
			"'FoldedStyle', 'FlowStyle'.")
	c.Flags().BoolVar(&r.StripComments, "strip-comments", false,
		"remove comments from yaml.")
	c.Flags().BoolVar(&r.IncludeLocal, "include-local", false,
		"if true, include local-config in the output.")
	c.Flags().BoolVar(&r.ExcludeNonLocal, "exclude-non-local", false,
		"if true, exclude non-local-config in the output.")
	c.Flags().StringVar(&r.OutputDest, "dest", "",
		"if specified, write output to a file rather than stdout")
	r.Command = c
	return r
}

func CatCommand(name string) *cobra.Command {
	return GetCatRunner(name).Command
}

// CatRunner contains the run function
type CatRunner struct {
	IncludeSubpackages bool
	Format             bool
	KeepAnnotations    bool
	WrapKind           string
	WrapApiVersion     string
	FunctionConfig     string
	OutputDest         string
	Styles             []string
	StripComments      bool
	IncludeLocal       bool
	ExcludeNonLocal    bool
	Command            *cobra.Command
}

func (r *CatRunner) runE(c *cobra.Command, args []string) error {
	// if there is a function-config specified, emit it
	var functionConfig *yaml.RNode
	if r.FunctionConfig != "" {
		configs, err := kio.LocalPackageReader{PackagePath: r.FunctionConfig,
			OmitReaderAnnotations: !r.KeepAnnotations}.Read()
		if err != nil {
			return err
		}
		if len(configs) != 1 {
			return fmt.Errorf("expected exactly 1 functionConfig, found %d", len(configs))
		}
		functionConfig = configs[0]
	}

	var inputs []kio.Reader
	for _, a := range args {
		inputs = append(inputs, kio.LocalPackageReader{
			PackagePath:        a,
			IncludeSubpackages: r.IncludeSubpackages,
		})
	}
	if len(inputs) == 0 {
		inputs = append(inputs, &kio.ByteReader{Reader: c.InOrStdin()})
	}
	var fltr []kio.Filter
	// don't include reconcilers
	fltr = append(fltr, &filters.IsLocalConfig{
		IncludeLocalConfig:    r.IncludeLocal,
		ExcludeNonLocalConfig: r.ExcludeNonLocal,
	})
	if r.Format {
		fltr = append(fltr, filters.FormatFilter{})
	}
	if r.StripComments {
		fltr = append(fltr, filters.StripCommentsFilter{})
	}

	var out = c.OutOrStdout()
	if r.OutputDest != "" {
		o, err := os.Create(r.OutputDest)
		if err != nil {
			return handleError(c, errors.Wrap(err))
		}
		defer o.Close()
		out = o
	}

	// remove this annotation explicitly, the ByteWriter won't clear it by
	// default because it doesn't set it
	clear := []string{"config.kubernetes.io/path"}
	if r.KeepAnnotations {
		clear = nil
	}

	var outputs []kio.Writer
	outputs = append(outputs, kio.ByteWriter{
		Writer:                out,
		KeepReaderAnnotations: r.KeepAnnotations,
		WrappingKind:          r.WrapKind,
		WrappingAPIVersion:    r.WrapApiVersion,
		FunctionConfig:        functionConfig,
		Style:                 yaml.GetStyle(r.Styles...),
		ClearAnnotations:      clear,
	})

	return handleError(c, kio.Pipeline{Inputs: inputs, Filters: fltr, Outputs: outputs}.Execute())
}
