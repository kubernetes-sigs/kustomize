// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type CLIMode byte

const (
	StandaloneEnabled CLIMode = iota
	StandaloneDisabled
)

// Build returns a cobra.Command to run a function.
//
// The cobra.Command reads the input from STDIN, invokes the provided processor,
// and then writes the output to STDOUT.
//
// The cobra.Command has a boolean `--stack` flag to print stack traces on failure.
//
// By default, invoking the returned cobra.Command with arguments triggers "standalone" mode.
// In this mode:
// - The first argument must be the name of a file containing the FunctionConfig.
// - The remaining arguments must be filenames containing input resources for ResourceList.Items.
// - The argument "-", if present, will cause resources to be read from STDIN as well.
// The output will be a raw stream of resources (not wrapped in a List type).
// Example usage: `cat input1.yaml | go run main.go config.yaml input2.yaml input3.yaml -`
//
// If mode is `StandaloneDisabled`, all arguments are ignored, and STDIN must contain
// a Kubernetes List type. To pass a function config in this mode, use a ResourceList as the input.
// The output will be of the same type as the input (e.g. ResourceList).
// Example usage: `cat resource_list.yaml | go run main.go`
//
// By default, any error returned by the ResourceListProcessor will be printed to STDERR.
// Set noPrintError to true to suppress this.
func Build(p framework.ResourceListProcessor, mode CLIMode, noPrintError bool) *cobra.Command {
	cmd := cobra.Command{}

	var printStack bool
	cmd.Flags().BoolVar(&printStack, "stack", false, "print the stack trace on failure")
	cmd.Args = cobra.MinimumNArgs(0)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var readers []io.Reader
		rw := &kio.ByteReadWriter{
			Writer:                cmd.OutOrStdout(),
			KeepReaderAnnotations: true,
		}

		if len(args) > 0 && mode == StandaloneEnabled {
			// Don't keep the reader annotations if we are in standalone mode
			rw.KeepReaderAnnotations = false
			// Don't wrap the resources in a resourceList -- we are in
			// standalone mode and writing to stdout to be applied
			rw.NoWrap = true

			for i := range args {
				// the first argument is the resourceList.FunctionConfig
				if i == 0 {
					var err error
					if rw.FunctionConfig, err = functionConfigFromFile(args[0]); err != nil {
						return errors.Wrap(err)
					}
					continue
				}
				if args[i] == "-" {
					readers = append([]io.Reader{cmd.InOrStdin()}, readers...)
				} else {
					readers = append(readers, &deferredFileReader{path: args[i]})
				}
			}
		} else {
			// legacy kustomize plugin input style
			legacyPlugin := os.Getenv("KUSTOMIZE_PLUGIN_CONFIG_STRING")
			if legacyPlugin != "" && rw.FunctionConfig != nil {
				if err := yaml.Unmarshal([]byte(legacyPlugin), rw.FunctionConfig); err != nil {
					return err
				}
			}
			readers = append(readers, cmd.InOrStdin())
		}
		rw.Reader = io.MultiReader(readers...)

		err := framework.Execute(p, rw)
		if err != nil && !noPrintError {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v", err)
		}
		// print the stack if requested
		if s := errors.GetStack(err); printStack && s != "" {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), s)
		}
		return err
	}

	return &cmd
}

// AddGenerateDockerfile adds a "gen" subcommand to create a Dockerfile for building
// the function into a container image.
// The gen command takes one argument: the directory where the Dockerfile will be created.
//
//	go run main.go gen DIR/
func AddGenerateDockerfile(cmd *cobra.Command) {
	gen := &cobra.Command{
		Use:  "gen [DIR]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.WriteFile(filepath.Join(args[0], "Dockerfile"), []byte(`FROM public.ecr.aws/docker/library/golang:1.22.7-bullseye as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags '-w -s' -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=builder /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]
`), 0600); err != nil {
				return fmt.Errorf("%w", err)
			}
			return nil
		},
	}
	cmd.AddCommand(gen)
}

func functionConfigFromFile(file string) (*yaml.RNode, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "unable to read configuration file %q", file)
	}
	fc, err := yaml.Parse(string(b))
	if err != nil {
		return nil, errors.WrapPrefixf(err, "unable to parse configuration file %q", file)
	}
	return fc, nil
}

type deferredFileReader struct {
	path      string
	srcReader io.Reader
}

func (fr *deferredFileReader) Read(dest []byte) (int, error) {
	if fr.srcReader == nil {
		src, err := os.ReadFile(fr.path)
		if err != nil {
			return 0, errors.WrapPrefixf(err, "unable to read input file %s", fr.path)
		}
		fr.srcReader = bytes.NewReader(src)
	}
	return fr.srcReader.Read(dest)
}
