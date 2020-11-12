// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/testutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCommand_dockerfile(t *testing.T) {
	d, err := ioutil.TempDir("", "kustomize")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(d)

	// create a function

	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error { return nil })

	// generate the Dockerfile
	cmd.SetArgs([]string{"gen", d})
	if !assert.NoError(t, cmd.Execute()) {
		t.FailNow()
	}

	b, err := ioutil.ReadFile(filepath.Join(d, "Dockerfile"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := `FROM golang:1.13-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go build -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
CMD ["function"]
`
	if !assert.Equal(t, expected, string(b)) {
		t.FailNow()
	}
}

// TestCommand_standalone tests the framework works in standalone mode
func TestCommand_standalone(t *testing.T) {
	// TODO: make this test pass on windows -- currently failure seems spurious
	testutil.SkipWindows(t)

	type api = struct {
		A string `json:"a" yaml:"a"`
	}
	var config api

	resourceList := &framework.ResourceList{FunctionConfig: &config}
	cmd := framework.Command(resourceList, func() error {
		resourceList.Items = append(resourceList.Items, yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar1
  namespace: default
  annotations:
    foo: bar1
`))

		for i := range resourceList.Items {
			err := resourceList.Items[i].PipeE(yaml.SetAnnotation("a", config.A))
			if err != nil {
				return err
			}
		}

		return nil
	})
	cmd.SetArgs([]string{
		filepath.Join("testdata", "command", "config.yaml"),
		filepath.Join("testdata", "command", "input.yaml"),
	})
	var out bytes.Buffer
	cmd.SetOutput(&out)
	if !assert.NoError(t, cmd.Execute()) {
		t.FailNow()
	}

	expected, err := ioutil.ReadFile(filepath.Join("testdata", "command", "expected.yaml"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, string(expected), out.String()) {
		t.FailNow()
	}
}
