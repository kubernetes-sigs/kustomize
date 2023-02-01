// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/provenance"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Options struct {
	Short  bool
	Output string
	Writer io.Writer
}

// NewCmdVersion makes a new version command.
func NewCmdVersion(w io.Writer) *cobra.Command {
	o := NewOptions(w)
	versionCmd := cobra.Command{
		Use:     "version",
		Short:   "Prints the kustomize version",
		Example: `kustomize version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	versionCmd.Flags().BoolVar(&o.Short, "short", false, "short form")
	_ = versionCmd.Flags().MarkDeprecated("short", "and will be removed in the future.")
	versionCmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'.")
	return &versionCmd
}

func NewOptions(w io.Writer) *Options {
	if w == nil {
		w = io.Writer(os.Stdout)
	}
	return &Options{Writer: w}
}

func (o *Options) Validate(_ []string) error {
	if o.Short {
		if o.Output != "" {
			return fmt.Errorf("--short and --output are mutually exclusive")
		}
	}
	return nil
}

func (o *Options) Run() error {
	switch o.Output {
	case "":
		if o.Short {
			fmt.Fprintln(o.Writer, provenance.GetProvenance().Short())
		} else {
			fmt.Fprintln(o.Writer, provenance.GetProvenance().Semver())
		}
	case "yaml":
		marshalled, err := yaml.Marshal(provenance.GetProvenance())
		if err != nil {
			return errors.WrapPrefixf(err, "marshalling provenance to yaml")
		}
		fmt.Fprintln(o.Writer, string(marshalled))
	case "json":
		marshalled, err := json.MarshalIndent(provenance.GetProvenance(), "", "  ")
		if err != nil {
			return errors.WrapPrefixf(err, "marshalling provenance to json")
		}
		fmt.Fprintln(o.Writer, string(marshalled))
	}
	return nil
}
