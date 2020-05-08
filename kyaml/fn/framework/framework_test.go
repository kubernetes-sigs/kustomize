// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
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
