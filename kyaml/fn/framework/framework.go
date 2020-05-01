// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Command provides a cobra.Command for running the function.
//
// If functionConfig is nil, the function may be configured with flags parsed from
// the ResourceList.functionConfig by creating flags on the returned command.
func Command(functionConfig interface{}, function Function) cobra.Command {
	cmd := cobra.Command{}
	addGenerate(&cmd)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := execute(function, functionConfig, cmd)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%v", err)
		}
		return err
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd
}

func addGenerate(cmd *cobra.Command) {
	gen := &cobra.Command{
		Use:  "gen",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ioutil.WriteFile(filepath.Join(args[0], "Dockerfile"), []byte(`FROM golang:1.13-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go build -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
CMD ["function"]
`), 0600)
		},
	}
	cmd.AddCommand(gen)
}

func execute(function Function, functionConfig interface{}, cmd *cobra.Command) error {
	rw := &kio.ByteReadWriter{
		Reader:                cmd.InOrStdin(),
		Writer:                cmd.OutOrStdout(),
		KeepReaderAnnotations: true,
	}
	nodes, err := rw.Read()
	if err != nil {
		return errors.Wrap(err)
	}

	// parse the functionConfig
	if rw.FunctionConfig != nil {
		if functionConfig == nil {
			functionConfig = map[string]interface{}{}
		}

		// unmarshal into the provided structure
		err := yaml.Unmarshal([]byte(rw.FunctionConfig.MustString()), functionConfig)
		if err != nil {
			return errors.Wrap(err)
		}

		// set the functionConfig values as flags so they are easy to access
		err = func() error {
			if !cmd.HasFlags() {
				return nil
			}
			// kpt serializes function arguments as a ConfigMap, read them from
			// the data field.
			fc, ok := functionConfig.(map[string]interface{})
			if !ok {
				// serialized as something else
				return nil
			}
			if fc["data"] == nil {
				return nil
			}
			data := fc["data"].(map[string]interface{})
			// set the value of each flag from the ResourceList.function config input
			// values
			for k, v := range data {
				s, ok := v.(string)
				if !ok {
					continue
				}
				if err = cmd.Flag(k).Value.Set(s); err != nil {
					return errors.Wrap(err)
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}

	// run the function implementation
	nodes, err = function(nodes)

	// set the ResourceList.results for validating functions
	var result *Result
	if err != nil {
		if val, ok := err.(Result); ok {
			if len(val.Items) > 0 {
				result = &val
				b, err := yaml.Marshal(val)
				if err != nil {
					return errors.Wrap(err)
				}
				y, err := yaml.Parse(string(b))
				if err != nil {
					return errors.Wrap(err)
				}
				rw.Results = y
			}
		} else {
			return errors.Wrap(err)
		}
	}

	// write the results
	if err := rw.Write(nodes); err != nil {
		return errors.Wrap(err)
	}

	if result != nil && result.ExitCode() != 0 {
		return errors.Wrap(err)
	}
	return nil
}
