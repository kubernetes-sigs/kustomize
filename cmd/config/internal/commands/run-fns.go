// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/runfn"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GetCatRunner returns a RunFnRunner.
func GetRunFnRunner(name string) *RunFnRunner {
	r := &RunFnRunner{}
	c := &cobra.Command{
		Use:     "run [DIR]",
		Short:   commands.RunFnsShort,
		Long:    commands.RunFnsLong,
		Example: commands.RunFnsExamples,
		RunE:    r.runE,
		PreRunE: r.preRunE,
	}
	fixDocs(name, c)
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	r.Command = c
	r.Command.Flags().BoolVar(
		&r.DryRun, "dry-run", false, "print results to stdout")
	r.Command.Flags().BoolVar(
		&r.GlobalScope, "global-scope", false, "set global scope for functions.")
	r.Command.Flags().StringSliceVar(
		&r.FnPaths, "fn-path", []string{},
		"read functions from these directories instead of the configuration directory.")
	r.Command.Flags().StringVar(
		&r.Image, "image", "",
		"run this image as a function instead of discovering them.")
	return r
}

func RunFnCommand(name string) *cobra.Command {
	return GetRunFnRunner(name).Command
}

// RunFnRunner contains the run function
type RunFnRunner struct {
	IncludeSubpackages bool
	Command            *cobra.Command
	DryRun             bool
	GlobalScope        bool
	FnPaths            []string
	Image              string
	RunFns             runfn.RunFns
}

func (r *RunFnRunner) runE(c *cobra.Command, args []string) error {
	return handleError(c, r.RunFns.Execute())
}

// getFunctions parses the commandline flags and arguments into explicit
// Functions to run.
func (r *RunFnRunner) getFunctions(c *cobra.Command, args, dataItems []string) (
	[]*yaml.RNode, error) {
	// if image isn't specified, then Functions is empty
	if r.Image == "" {
		return nil, nil
	}

	// create the function spec to set as an annotation
	fn, err := yaml.Parse(`container: {}`)
	if err != nil {
		return nil, err
	}
	// TODO: add support network, volumes, etc based on flag values
	err = fn.PipeE(
		yaml.Lookup("container"),
		yaml.SetField("image", yaml.NewScalarRNode(r.Image)))
	if err != nil {
		return nil, err
	}

	// create the function config
	rc, err := yaml.Parse(`
metadata:
  name: function-input
data: {}
`)
	if err != nil {
		return nil, err
	}

	// set the function annotation on the function config so it
	// is parsed by RunFns
	value, err := fn.String()
	if err != nil {
		return nil, err
	}
	err = rc.PipeE(
		yaml.LookupCreate(yaml.MappingNode, "metadata", "annotations"),
		yaml.SetField("config.kubernetes.io/function", yaml.NewScalarRNode(value)))
	if err != nil {
		return nil, err
	}

	// default the function config kind to ConfigMap, this may be overridden
	var kind = "ConfigMap"
	var version = "v1"

	// populate the function config with data.  this is a convention for functions
	// to be more commandline friendly
	if len(dataItems) > 0 {
		dataField, err := rc.Pipe(yaml.Lookup("data"))
		if err != nil {
			return nil, err
		}
		for i, s := range dataItems {
			kv := strings.SplitN(s, "=", 2)
			if i == 0 && len(kv) == 1 {
				// first argument may be the kind
				kind = s
				continue
			}
			if len(kv) != 2 {
				return nil, fmt.Errorf("args must have keys and values separated by =")
			}
			err := dataField.PipeE(yaml.SetField(kv[0], yaml.NewScalarRNode(kv[1])))
			if err != nil {
				return nil, err
			}
		}
	}
	err = rc.PipeE(yaml.SetField("kind", yaml.NewScalarRNode(kind)))
	if err != nil {
		return nil, err
	}
	err = rc.PipeE(yaml.SetField("apiVersion", yaml.NewScalarRNode(version)))
	if err != nil {
		return nil, err
	}
	return []*yaml.RNode{rc}, nil
}

func (r *RunFnRunner) preRunE(c *cobra.Command, args []string) error {
	if c.ArgsLenAtDash() >= 0 && r.Image == "" {
		return errors.Errorf("must specify --image")
	}

	var dataItems []string
	if c.ArgsLenAtDash() >= 0 {
		dataItems = args[c.ArgsLenAtDash():]
		args = args[:c.ArgsLenAtDash()]
	}
	if len(args) > 1 {
		return errors.Errorf("0 or 1 arguments supported, function arguments go after '--'")
	}

	fns, err := r.getFunctions(c, args, dataItems)
	if err != nil {
		return err
	}

	// set the output to stdout if in dry-run mode or no arguments are specified
	var output io.Writer
	var input io.Reader
	if len(args) == 0 {
		output = c.OutOrStdout()
		input = c.InOrStdin()
	} else if r.DryRun {
		output = c.OutOrStdout()
	}

	// set the path if specified as an argument
	var path string
	if len(args) == 1 {
		// argument is the directory
		path = args[0]
	}

	r.RunFns = runfn.RunFns{
		FunctionPaths: r.FnPaths,
		GlobalScope:   r.GlobalScope,
		Functions:     fns,
		Output:        output,
		Input:         input,
		Path:          path,
	}

	// don't consider args for the function
	return nil
}
