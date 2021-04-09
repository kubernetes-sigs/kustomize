// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"testing"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestResourcePatchTemplate_ComplexSelectors(t *testing.T) {
	cmdFn := func() *cobra.Command {
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
			ResourceMatcher: func(rn *yaml.RNode) bool {
				m, _ := rn.GetMeta()
				return config.Special != "" && m.Annotations["foo"] == config.Special
			},
		}
		pt1 := framework.ResourcePatchTemplate{
			// Apply these rendered patches
			Templates: parser.TemplateStrings(`
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
`),
			// Use the selector from the input
			Selector: &config.Selector,
		}

		pt2 := framework.ResourcePatchTemplate{
			// Apply these rendered patches
			Templates: parser.TemplateStrings(`
metadata:
  annotations:
    filterPatched: '{{ .A }}'
`),
			// Use an explicit selector
			Selector: &filter,
		}

		fn := framework.TemplateProcessor{
			TemplateData: &config,
			PreProcessFilters: []kio.Filter{kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
				// do some extra processing based on the inputs
				config.LongList = len(items) > 2
				return items, nil
			})},
			PatchTemplates: []framework.PatchTemplate{&pt1, &pt2},
		}

		return command.Build(fn, command.StandaloneEnabled, false)
	}

	tc := frameworktestutil.CommandResultsChecker{Command: cmdFn,
		TestDataDirectory: "testdata/patch-selector"}
	tc.Assert(t)
}
