// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package commands

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands/internal/k8sgen/pkg/api/resource"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/cmd/config/runner"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

// Cmd returns a command GrepRunner.
func GetGrepRunner(name string) *GrepRunner {
	r := &GrepRunner{}
	c := &cobra.Command{
		Use:     "grep QUERY [DIR]",
		Short:   commands.GrepShort,
		Long:    commands.GrepLong,
		Example: commands.GrepExamples,
		PreRunE: r.preRunE,
		RunE:    r.runE,
		Args:    cobra.MaximumNArgs(2),
	}
	runner.FixDocs(name, c)
	c.Flags().BoolVar(&r.KeepAnnotations, "annotate", true,
		"annotate resources with their file origins.")
	c.Flags().BoolVarP(&r.InvertMatch, "invert-match", "", false,
		"Selected Resources are those not matching any of the specified patterns..")
	c.Flags().BoolVarP(&r.RecurseSubPackages, "recurse-subpackages", "R", true,
		"also print resources recursively in all the nested subpackages")
	r.Command = c
	return r
}

func GrepCommand(name string) *cobra.Command {
	return GetGrepRunner(name).Command
}

// GrepRunner contains the run function
type GrepRunner struct {
	KeepAnnotations bool
	Command         *cobra.Command
	filters.GrepFilter
	Format             bool
	RecurseSubPackages bool
}

func (r *GrepRunner) preRunE(c *cobra.Command, args []string) error {
	r.GrepFilter.Compare = func(a, b string) (int, error) {
		qa, err := resource.ParseQuantity(a)
		if err != nil {
			return 0, fmt.Errorf("%s: %v", a, err)
		}
		qb, err := resource.ParseQuantity(b)
		if err != nil {
			return 0, err
		}

		return qa.Cmp(qb), err
	}
	parts, err := runner.ParseFieldPath(args[0])
	if err != nil {
		return err
	}

	var last []string
	if strings.Contains(parts[len(parts)-1], ">=") {
		last = strings.Split(parts[len(parts)-1], ">=")
		r.MatchType = filters.GreaterThanEq
	} else if strings.Contains(parts[len(parts)-1], "<=") {
		last = strings.Split(parts[len(parts)-1], "<=")
		r.MatchType = filters.LessThanEq
	} else if strings.Contains(parts[len(parts)-1], ">") {
		last = strings.Split(parts[len(parts)-1], ">")
		r.MatchType = filters.GreaterThan
	} else if strings.Contains(parts[len(parts)-1], "<") {
		last = strings.Split(parts[len(parts)-1], "<")
		r.MatchType = filters.LessThan
	} else {
		last = strings.Split(parts[len(parts)-1], "=")
		r.MatchType = filters.Regexp
	}
	if len(last) > 2 {
		return fmt.Errorf(
			"ambiguous match -- multiple of ['<', '>', '<=', '>=', '=' in final path element: %s",
			parts[len(parts)-1])
	}

	if len(last) > 1 {
		r.Value = last[1]
	}

	r.Path = append(parts[:len(parts)-1], last[0])
	return nil
}

func (r *GrepRunner) runE(c *cobra.Command, args []string) error {
	if len(args) == 1 {
		input := &kio.ByteReader{Reader: c.InOrStdin()}
		return runner.HandleError(c, kio.Pipeline{
			Inputs:  []kio.Reader{input},
			Filters: []kio.Filter{r.GrepFilter},
			Outputs: []kio.Writer{kio.ByteWriter{
				Writer:                c.OutOrStdout(),
				KeepReaderAnnotations: r.KeepAnnotations,
			}},
		}.Execute())
	}

	out := bytes.Buffer{}

	e := runner.ExecuteCmdOnPkgs{
		Writer:             &out,
		NeedOpenAPI:        false,
		RecurseSubPackages: r.RecurseSubPackages,
		CmdRunner:          r,
		RootPkgPath:        args[1],
		SkipPkgPathPrint:   true,
	}

	err := e.Execute()
	if err != nil {
		return err
	}

	res := strings.TrimSuffix(out.String(), "---")
	fmt.Fprintf(c.OutOrStdout(), "%s", res)

	return nil

}

func (r *GrepRunner) ExecuteCmd(w io.Writer, pkgPath string) error {
	input := kio.LocalPackageReader{PackagePath: pkgPath, PackageFileName: ext.KRMFileName()}
	out := &bytes.Buffer{}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{input},
		Filters: []kio.Filter{r.GrepFilter},
		Outputs: []kio.Writer{kio.ByteWriter{
			Writer:                out,
			KeepReaderAnnotations: r.KeepAnnotations,
		}},
	}.Execute()

	if err != nil {
		// return err if there is only package
		if !r.RecurseSubPackages {
			return err
		}
		// print error message and continue if there are multiple packages to annotate
		fmt.Fprintf(w, "%s\n", err.Error())
	}
	fmt.Fprint(w, out.String())
	if out.String() != "" {
		fmt.Fprint(w, "---")
	}
	return nil
}
