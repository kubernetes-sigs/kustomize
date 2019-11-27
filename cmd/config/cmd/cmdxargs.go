// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GetXArgsRunner returns a command runner.
func GetXArgsRunner() *XArgsRunner {
	r := &XArgsRunner{}
	c := &cobra.Command{
		Use:   "xargs -- CMD...",
		Short: "Convert functionConfig to commandline flags and envs",
		Long: `Convert functionConfig to commandline flags and envs.

xargs reads a ResourceList from stdin and parses the functionConfig field.  xargs then
reads each of the fields under .spec and parses them as flags.  If the fields have non-scalar
values, then xargs encoded the values as yaml strings.

  CMD:
    The command to run and pass the functionConfig as arguments.
`,
		Example: `
# given this example functionConfig in config.yaml
kind: Foo
spec:
  flag1: value1
  flag2: value2
items:
- 2
- 1

# this command:
$ kyaml cat pkg/ --function-config config.yaml --wrap-kind ResourceList | kyaml run xargs -- app

# is equivalent to this command:
$ kyaml cat pkg/ --function-config config.yaml --wrap-kind ResourceList | app --flag1=value1 --flag2=value2 2 1

# echo: prints the app arguments
$ kyaml cat pkg/ --function-config config.yaml --wrap-kind ResourceList | kyaml reconcile xargs -- echo
--flag1=value1 --flag2=value2 2 1

# env: prints the app env
$ kyaml cat pkg/ --function-config config.yaml --wrap-kind ResourceList | kyaml reconcile xargs -- env

# cat: prints the app stdin -- prints the package contents and functionConfig wrapped in a
# ResourceList
$ kyaml cat pkg/ --function-config config.yaml --wrap-kind ResourceList | kyaml reconcile xargs --no-flags -- env

`,
		RunE:               r.runE,
		SilenceUsage:       true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Args:               cobra.MinimumNArgs(1),
	}
	r.Command = c
	r.Command.Flags().BoolVar(&r.EnvOnly, "env-only", false, "only add env vars, not flags")
	c.Flags().StringVar(&r.WrapKind, "wrap-kind", "List", "wrap the input xargs give to the command in this type.")
	c.Flags().StringVar(&r.WrapVersion, "wrap-version", "v1", "wrap the input xargs give to the command in this type.")
	return r
}

// Runner contains the run function
type XArgsRunner struct {
	Command     *cobra.Command
	Args        []string
	EnvOnly     bool
	WrapKind    string
	WrapVersion string
}

func XArgsCommand() *cobra.Command {
	return GetXArgsRunner().Command
}

func (r *XArgsRunner) runE(c *cobra.Command, _ []string) error {
	if len(r.Args) == 0 {
		r.Args = os.Args
	}
	cmdIndex := -1
	for i := range r.Args {
		if r.Args[i] == "--" {
			cmdIndex = i + 1
			break
		}
	}
	if cmdIndex < 0 {
		return fmt.Errorf("must specify -- before command")
	}
	r.Args = r.Args[cmdIndex:]
	run := exec.Command(r.Args[0])

	if len(r.Args) > 1 {
		r.Args = r.Args[cmdIndex+1:]
	} else {
		r.Args = []string{}
	}
	run.Stdout = c.OutOrStdout()
	run.Stderr = c.ErrOrStderr()

	rw := &kio.ByteReadWriter{
		Reader: c.InOrStdin(),
	}
	nodes, err := rw.Read()
	if err != nil {
		return err
	}

	env := os.Environ()

	// append the config to the flags
	if err = func() error {
		if rw.FunctionConfig == nil {
			return nil
		}
		str, err := rw.FunctionConfig.String()
		if err != nil {
			return err
		}
		// add the API object to the env
		env = append(env, fmt.Sprintf("KUST_FUNCTION_CONFIG=%s", str))

		// parse the fields
		meta := rw.FunctionConfig.Field("metadata")
		if meta != nil {
			err = meta.Value.VisitFields(func(node *yaml.MapNode) error {
				if !r.EnvOnly {
					r.Args = append(r.Args, fmt.Sprintf("--%s=%s",
						node.Key.YNode().Value, parseYNode(node.Value.YNode())))
				}
				env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(node.Key.YNode().Value),
					node.Value.YNode().Value))
				return nil
			})
			if err != nil {
				return err
			}
		}

		spec := rw.FunctionConfig.Field("spec")
		if spec != nil {
			err = spec.Value.VisitFields(func(node *yaml.MapNode) error {
				if !r.EnvOnly {
					r.Args = append(r.Args, fmt.Sprintf("--%s=%s",
						node.Key.YNode().Value, parseYNode(node.Value.YNode())))
				}
				env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(node.Key.YNode().Value),
					node.Value.YNode().Value))
				return nil
			})
			if err != nil {
				return err
			}
		}

		if !r.EnvOnly {
			items := rw.FunctionConfig.Field("items")
			if items != nil {
				err = items.Value.VisitElements(func(node *yaml.RNode) error {
					r.Args = append(r.Args, parseYNode(node.YNode()))
					return nil
				})
				if err != nil {
					return err
				}
			}
		}

		if r.WrapKind != "" {
			if kind := rw.FunctionConfig.Field("kind"); !yaml.IsFieldEmpty(kind) {
				kind.Value.YNode().Value = r.WrapKind
			}
			rw.WrappingKind = r.WrapKind
		}
		if r.WrapVersion != "" {
			if version := rw.FunctionConfig.Field("apiVersion"); !yaml.IsFieldEmpty(version) {
				version.Value.YNode().Value = r.WrapVersion
			}
			rw.WrappingApiVersion = r.WrapVersion
		}
		return nil
	}(); err != nil {
		return err
	}
	run.Args = append(run.Args, r.Args...)
	run.Env = append(run.Env, env...)

	// write ResourceList to stdin
	if err = func() error {
		in, err := run.StdinPipe()
		if err != nil {
			return err
		}
		defer in.Close()
		rw.Writer = in
		if r.WrapKind != kio.ResourceListKind {
			rw.FunctionConfig = nil
		}
		return rw.Write(nodes)
	}(); err != nil {
		return err
	}

	return run.Run()
}

func parseYNode(node *yaml.Node) string {
	node.Value = strings.TrimSpace(node.Value)
	for _, b := range node.Value {
		if unicode.IsSpace(b) {
			// wrap in '' -- contains whitespace
			return fmt.Sprintf("'%s'", node.Value)
		}
	}
	return node.Value
}
