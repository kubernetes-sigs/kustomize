// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"text/template"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

func main() {
	type api struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
	}
	cmd := framework.TemplateCommand{
		API: &api{},
		Template: template.Must(template.New("example").Parse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: default
  annotations:
    {{ .Key }}: {{ .Value }}
`)),
	}.GetCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(cmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
