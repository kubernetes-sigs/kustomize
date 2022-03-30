// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/runner"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/runfn"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
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
	runner.FixDocs(name, c)
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
	// NOTE: exec plugins execute arbitrary code -- never change the default value of this flag!!!
	r.Command.Flags().BoolVar(
		&r.EnableExec, "enable-exec", false, /*do not change!*/
		"enable support for exec functions -- note: exec functions run arbitrary code -- do not use for untrusted configs!!! (Alpha)")
	r.Command.Flags().StringVar(
		&r.ExecPath, "exec-path", "", "run an executable as a function. (Alpha)")
	r.Command.Flags().BoolVar(
		&r.EnableStar, "enable-star", false, "enable support for starlark functions. (Alpha)")
	r.Command.Flags().StringVar(
		&r.StarPath, "star-path", "", "run a starlark script as a function. (Alpha)")
	r.Command.Flags().StringVar(
		&r.StarURL, "star-url", "", "run a starlark script as a function. (Alpha)")
	r.Command.Flags().StringVar(
		&r.StarName, "star-name", "", "name of starlark program. (Alpha)")

	r.Command.Flags().StringVar(
		&r.ResultsDir, "results-dir", "", "write function results to this dir")

	r.Command.Flags().BoolVar(
		&r.Network, "network", false, "enable network access for functions that declare it")
	r.Command.Flags().StringArrayVar(
		&r.Mounts, "mount", []string{},
		"a list of storage options read from the filesystem")
	r.Command.Flags().BoolVar(
		&r.LogSteps, "log-steps", false, "log steps to stderr")
	r.Command.Flags().StringArrayVarP(
		&r.Env, "env", "e", []string{},
		"a list of environment variables to be used by functions")
	r.Command.Flags().BoolVar(
		&r.AsCurrentUser, "as-current-user", false, "use the uid and gid of the command executor to run the function in the container")

	return r
}

func RunCommand(name string) *cobra.Command {
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
	EnableStar         bool
	StarPath           string
	StarURL            string
	StarName           string
	EnableExec         bool
	ExecPath           string
	RunFns             runfn.RunFns
	ResultsDir         string
	Network            bool
	Mounts             []string
	LogSteps           bool
	Env                []string
	AsCurrentUser      bool
}

func (r *RunFnRunner) runE(c *cobra.Command, args []string) error {
	return runner.HandleError(c, r.RunFns.Execute())
}

// getContainerFunctions parses the commandline flags and arguments into explicit
// Functions to run.
func (r *RunFnRunner) getContainerFunctions(dataItems []string) (
	[]*yaml.RNode, error) {
	if r.Image == "" && r.StarPath == "" && r.ExecPath == "" && r.StarURL == "" {
		return nil, nil
	}

	res, err := buildFnConfigResource(dataItems)
	if err != nil {
		return nil, err
	}

	// create the function spec to set as an annotation
	var fnAnnotation *yaml.RNode
	switch {
	case r.Image != "":
		fnAnnotation, err = fnAnnotationForImage(r.Image, r.Network)
	case r.EnableStar && (r.StarPath != "" || r.StarURL != ""):
		fnAnnotation, err = fnAnnotationForStar(r.StarPath, r.StarURL, r.StarName)
	case r.EnableExec && r.ExecPath != "":
		fnAnnotation, err = fnAnnotationForExec(r.ExecPath)
	}
	if err != nil {
		return nil, err
	}

	// set the function annotation on the function config, so that it is parsed by RunFns
	value, err := fnAnnotation.String()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	err = res.PipeE(
		yaml.LookupCreate(yaml.MappingNode, "metadata", "annotations"),
		yaml.SetField(runtimeutil.FunctionAnnotationKey, yaml.NewScalarRNode(value)))
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return []*yaml.RNode{res}, nil
}

func buildFnConfigResource(dataItems []string) (*yaml.RNode, error) {
	// create the function config
	rc, err := yaml.Parse(`
metadata:
  name: function-input
data: {}
`)
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
			// do not set style since function should determine tag and style
			err := dataField.PipeE(
				yaml.FieldSetter{Name: kv[0], Value: yaml.NewScalarRNode(kv[1]), OverrideStyle: true})
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
	return rc, nil
}

func fnAnnotationForExec(path string) (*yaml.RNode, error) {
	fn, err := yaml.Parse(`exec: {}`)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	err = fn.PipeE(
		yaml.Lookup("exec"),
		yaml.SetField("path", yaml.NewScalarRNode(path)))
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return fn, nil
}

func fnAnnotationForStar(path string, url string, name string) (*yaml.RNode, error) {
	fn, err := yaml.Parse(`starlark: {}`)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if path != "" {
		err = fn.PipeE(
			yaml.Lookup("starlark"),
			yaml.SetField("path", yaml.NewScalarRNode(path)))
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	if url != "" {
		err = fn.PipeE(
			yaml.Lookup("starlark"),
			yaml.SetField("url", yaml.NewScalarRNode(url)))
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	err = fn.PipeE(
		yaml.Lookup("starlark"),
		yaml.SetField("name", yaml.NewScalarRNode(name)))
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return fn, nil
}

func fnAnnotationForImage(image string, enableNetwork bool) (*yaml.RNode, error) {
	fn, err := yaml.Parse(`container: {}`)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	// TODO: add support network, volumes, etc based on flag values
	err = fn.PipeE(
		yaml.Lookup("container"),
		yaml.SetField("image", yaml.NewScalarRNode(image)))
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if enableNetwork {
		n := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "true",
			Tag:   yaml.NodeTagBool,
		}
		err = fn.PipeE(
			yaml.Lookup("container"),
			yaml.SetField("network", yaml.NewRNode(n)))
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	return fn, nil
}

func toStorageMounts(mounts []string) []runtimeutil.StorageMount {
	var sms []runtimeutil.StorageMount
	for _, mount := range mounts {
		sms = append(sms, runtimeutil.StringToStorageMount(mount))
	}
	return sms
}

func (r *RunFnRunner) preRunE(c *cobra.Command, args []string) error {
	if !r.EnableStar && (r.StarPath != "" || r.StarURL != "") {
		return errors.Errorf("must specify --enable-star with --star-path and --star-url")
	}

	if !r.EnableExec && r.ExecPath != "" {
		return errors.Errorf("must specify --enable-exec with --exec-path")
	}

	if c.ArgsLenAtDash() >= 0 && r.Image == "" &&
		!(r.EnableStar && (r.StarPath != "" || r.StarURL != "")) && !(r.EnableExec && r.ExecPath != "") {
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

	fns, err := r.getContainerFunctions(dataItems)
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

	// parse mounts to set storageMounts
	storageMounts := toStorageMounts(r.Mounts)

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	r.RunFns = runfn.RunFns{
		FunctionPaths:  r.FnPaths,
		GlobalScope:    r.GlobalScope,
		Functions:      fns,
		Output:         output,
		Input:          input,
		Path:           path,
		Network:        r.Network,
		EnableStarlark: r.EnableStar,
		EnableExec:     r.EnableExec,
		StorageMounts:  storageMounts,
		ResultsDir:     r.ResultsDir,
		LogSteps:       r.LogSteps,
		Env:            r.Env,
		AsCurrentUser:  r.AsCurrentUser,
		WorkingDir:     wd,
	}

	// don't consider args for the function
	return nil
}
