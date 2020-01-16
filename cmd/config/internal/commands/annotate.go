// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// NewAnnotateRunner returns a command runner.
func NewAnnotateRunner(parent string) *AnnotateRunner {
	r := &AnnotateRunner{}
	c := &cobra.Command{
		Use:     "annotate [DIR]",
		Args:    cobra.MaximumNArgs(1),
		Short:   commands.AnnotateShort,
		Long:    commands.AnnotateLong,
		Example: commands.AnnotateExamples,
		RunE:    r.runE,
	}
	fixDocs(parent, c)
	r.Command = c
	c.Flags().StringVar(&r.Kind, "kind", "", "Resource kind to annotate")
	c.Flags().StringVar(&r.ApiVersion, "apiVersion", "", "Resource apiVersion to annotate")
	c.Flags().StringVar(&r.Name, "name", "", "Resource name to annotate")
	c.Flags().StringVar(&r.Namespace, "namespace", "", "Resource namespace to annotate")
	c.Flags().StringSliceVar(&r.Values, "kv", []string{}, "annotation as KEY=VALUE")
	return r
}

func AnnotateCommand(parent string) *cobra.Command {
	return NewAnnotateRunner(parent).Command
}

type AnnotateRunner struct {
	Command    *cobra.Command
	Values     []string
	Kind       string
	Name       string
	ApiVersion string
	Namespace  string
	Path       string
}

func (r *AnnotateRunner) runE(c *cobra.Command, args []string) error {
	var input []kio.Reader
	var output []kio.Writer
	if len(args) == 0 {
		rw := &kio.ByteReadWriter{Reader: c.InOrStdin(), Writer: c.OutOrStdout()}
		input = []kio.Reader{rw}
		output = []kio.Writer{rw}
	} else {
		rw := &kio.LocalPackageReadWriter{PackagePath: args[0], NoDeleteFiles: true}
		input = []kio.Reader{rw}
		output = []kio.Writer{rw}
	}
	return handleError(c, kio.Pipeline{
		Inputs:  input,
		Filters: []kio.Filter{r},
		Outputs: output,
	}.Execute())
}

func (r *AnnotateRunner) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range nodes {
		n := nodes[i]
		m, err := n.GetMeta()
		if err != nil {
			return nil, err
		}
		if r.Kind != "" && r.Kind != m.Kind {
			continue
		}
		if r.ApiVersion != "" && r.ApiVersion != m.APIVersion {
			continue
		}
		if r.Namespace != "" && r.Namespace != m.Namespace {
			continue
		}
		if r.Name != "" && r.Name != m.Name {
			continue
		}

		for i := range r.Values {
			// split key, value pairs
			kv := strings.SplitN(r.Values[i], "=", 2)
			if len(kv) != 2 {
				return nil, errors.Errorf("must specify --kv as KEY=VALUE: %s", r.Values[i])
			}
			if err := n.PipeE(yaml.SetAnnotation(kv[0], kv[1])); err != nil {
				return nil, err
			}
		}

	}
	return nodes, nil
}
