// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package command_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCommand_dockerfile(t *testing.T) {
	d := t.TempDir()

	// create a function
	cmd := command.Build(&framework.SimpleProcessor{}, command.StandaloneEnabled, false)
	// add the Dockerfile generator
	command.AddGenerateDockerfile(cmd)

	// generate the Dockerfile
	cmd.SetArgs([]string{"gen", d})
	if !assert.NoError(t, cmd.Execute()) {
		t.FailNow()
	}

	b, err := os.ReadFile(filepath.Join(d, "Dockerfile"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := `FROM golang:1.18-alpine as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags '-w -s' -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=builder /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]
`
	if !assert.Equal(t, expected, string(b)) {
		t.FailNow()
	}
}

// TestCommand_standalone tests the framework works in standalone mode
func TestCommand_standalone(t *testing.T) {
	var config struct {
		A string `json:"a" yaml:"a"`
		B int    `json:"b" yaml:"b"`
	}

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		items = append(items, yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
 name: bar1
 namespace: default
 annotations:
   foo: bar1
`))
		for i := range items {
			err := items[i].PipeE(yaml.SetAnnotation("a", config.A))
			if err != nil {
				return nil, err
			}
			err = items[i].PipeE(yaml.SetAnnotation("b", fmt.Sprintf("%v", config.B)))
			if err != nil {
				return nil, err
			}
		}

		return items, nil
	}

	cmdFn := func() *cobra.Command {
		return command.Build(&framework.SimpleProcessor{Filter: kio.FilterFunc(fn), Config: &config}, command.StandaloneEnabled, false)
	}

	tc := frameworktestutil.CommandResultsChecker{Command: cmdFn}
	tc.Assert(t)
}

func TestCommand_standalone_stdin(t *testing.T) {
	var config struct {
		A string `json:"a" yaml:"a"`
		B int    `json:"b" yaml:"b"`
	}

	p := &framework.SimpleProcessor{
		Config: &config,

		Filter: kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
			items = append(items, yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar2
  namespace: default
  annotations:
    foo: bar2
`))
			for i := range items {
				err := items[i].PipeE(yaml.SetAnnotation("a", config.A))
				if err != nil {
					return nil, err
				}
				err = items[i].PipeE(yaml.SetAnnotation("b", fmt.Sprintf("%v", config.B)))
				if err != nil {
					return nil, err
				}
			}

			return items, nil
		}),
	}
	cmd := command.Build(p, command.StandaloneEnabled, false)
	cmd.SetIn(bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar1
  namespace: default
  annotations:
    foo: bar1
spec:
  replicas: 1
`))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{filepath.Join("testdata", "standalone", "config.yaml"), "-"})

	require.NoError(t, cmd.Execute())

	require.Equal(t, strings.TrimSpace(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar1
  namespace: default
  annotations:
    foo: bar1
    a: 'c'
    b: '1'
spec:
  replicas: 1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar2
  namespace: default
  annotations:
    foo: bar2
    a: 'c'
    b: '1'
`), strings.TrimSpace(out.String()))
}
