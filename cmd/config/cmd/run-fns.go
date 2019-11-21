// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/runfn"
)

// GetCatRunner returns a RunFnRunner.
func GetRunFnRunner() *RunFnRunner {
	r := &RunFnRunner{}
	c := &cobra.Command{
		Use:   "run-fns DIR",
		Short: "Apply config functions to Resources.",
		Long: `Apply config functions to Resources.

run-fns sequentially invokes all config functions in the directly, providing Resources
in the directory as input to the first function, and writing the output of the last
function back to the directory.

The ordering of functions is determined by the order they are encountered when walking the
directory.  To clearly specify an ordering of functions, multiple functions may be
declared in the same file, separated by '---' (the functions will be invoked in the
order they appear in the file).

### Arguments:

  DIR:
    Path to local directory.


### Config Functions:

  Config functions are specified as Kubernetes types containing a metadata.configFn.container.image
  field.  This fields tells run-fns how to invoke the container.

  Example config function:

	# in file example/fn.yaml
	apiVersion: fn.example.com/v1beta1
	kind: ExampleFunctionKind
	metadata:
	  configFn:
	    container:
	      # function is invoked as a container running this image
	      image: gcr.io/example/examplefunction:v1.0.1
	  annotations:
	    config.kubernetes.io/local-config: "true" # tools should ignore this
	spec:
	  configField: configValue

  In the preceding example, 'kyaml run-fns example/' would identify the function by
  the metadata.configFn field.  It would then write all Resources in the directory to
  a container stdin (running the gcr.io/example/examplefunction:v1.0.1 image).  It
  would then writer the container stdout back to example/, replacing the directory
  file contents.
`,
		Example: `
kyaml run-fns example/
`,
		RunE: r.runE,
		Args: cobra.ExactArgs(1),
	}
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	r.Command = c
	r.Command.Flags().BoolVar(
		&r.DryRun, "dry-run", false, "print results to stdout")
	r.Command.Flags().StringSliceVar(
		&r.FnPaths, "fn-path", []string{},
		"directories containing functions without configuration")
	r.Command.AddCommand(XArgsCommand())
	r.Command.AddCommand(WrapCommand())
	return r
}

func RunFnCommand() *cobra.Command {
	return GetRunFnRunner().Command
}

// RunFnRunner contains the run function
type RunFnRunner struct {
	IncludeSubpackages bool
	Command            *cobra.Command
	DryRun             bool
	FnPaths            []string
}

func (r *RunFnRunner) runE(c *cobra.Command, args []string) error {
	rec := runfn.RunFns{Path: args[0], FunctionPaths: r.FnPaths}
	if r.DryRun {
		rec.Output = c.OutOrStdout()
	}
	return rec.Execute()

}
