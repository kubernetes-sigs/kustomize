// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GetCatRunner returns a command CatRunner.
func GetCatRunner() *CatRunner {
	r := &CatRunner{}
	c := &cobra.Command{
		Use:   "cat DIR...",
		Short: "Print Resource Config from a local directory",
		Long: `Print Resource Config from a local directory.

  DIR:
    Path to local directory.
`,
		Example: `# print Resource config from a directory
kyaml cat my-dir/

# wrap Resource config from a directory in an ResourceList
kyaml cat my-dir/ --wrap-kind ResourceList --wrap-version config.kubernetes.io/v1alpha1 --function-config fn.yaml

# unwrap Resource config from a directory in an ResourceList
... | kyaml cat
`,
		RunE: r.runE,
	}
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
	r.Command = c
	return r
}

func CatCommand() *cobra.Command {
	return GetCatRunner().Command
}

// CatRunner contains the run function
type CatRunner struct {
	IncludeSubpackages bool
	Format             bool
	KeepAnnotations    bool
	WrapKind           string
	WrapApiVersion     string
	FunctionConfig     string
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

	var outputs []kio.Writer
	outputs = append(outputs, kio.ByteWriter{
		Writer:                c.OutOrStdout(),
		KeepReaderAnnotations: r.KeepAnnotations,
		WrappingKind:          r.WrapKind,
		WrappingApiVersion:    r.WrapApiVersion,
		FunctionConfig:        functionConfig,
		Style:                 yaml.GetStyle(r.Styles...),
	})

	return handleError(c, kio.Pipeline{Inputs: inputs, Filters: fltr, Outputs: outputs}.Execute())
}
