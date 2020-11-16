// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"testing"
	"text/template"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
	"sigs.k8s.io/kustomize/kyaml/testutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestPatchTemplate(t *testing.T) {
	// TODO: make this test pass on windows -- current failure seems spurious
	testutil.SkipWindows(t)

	cmdFn := func() cobra.Command {
		type api struct {
			Selector framework.Selector `json:"selector" yaml:"selector"`
			A        string             `json:"a" yaml:"a"`
			B        string             `json:"b" yaml:"b"`
			Special  string             `json:"special" yaml:"special"`
			LongList bool
		}
		var config api
		filter := framework.Selector{
			// this is a special manual filter for the Selector for when the built-in matchers
			// are insufficient
			Filter: func(rn *yaml.RNode) bool {
				m, _ := rn.GetMeta()
				return config.Special != "" && m.Annotations["foo"] == config.Special
			},
		}
		return framework.TemplateCommand{
			API: &config,
			PreProcess: func(rl *framework.ResourceList) error {
				// do some extra processing based on the inputs
				config.LongList = len(rl.Items) > 2
				return nil
			},
			PatchTemplates: []framework.PatchTemplate{
				{
					// Apply these rendered patches
					Template: template.Must(template.New("test").Parse(`
spec:
  template:
    spec:
      containers:
      - name: foo
        image: example/sidecar:{{ .B }}
---
metadata:
  annotations:
    patched: '{{ .A }}'
{{- if .LongList }}
    long: 'true'
{{- end }}
`)),
					// Use the selector from the input
					Selector: &config.Selector,
				},
				{
					// Apply these rendered patches
					Template: template.Must(template.New("test").Parse(`
metadata:
  annotations:
    filterPatched: '{{ .A }}'
`)),
					// Use an explicit selector
					Selector: &filter,
				},
			},
		}.GetCommand()
	}

	frameworktestutil.ResultsChecker{Command: cmdFn, TestDataDirectory: "patchtestdata"}.Assert(t)
}
