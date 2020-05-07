// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package exec_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/exec"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFunctionFilter_Filter(t *testing.T) {
	var tests = []struct {
		name           string
		input          []string
		functionConfig string
		expectedOutput []string
		expectedError  string
		instance       exec.Filter
	}{
		{
			name: "exec_sed",
			input: []string{
				`apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo`,
				`apiVersion: v1
kind: Service
metadata:
  name: service-foo`,
			},
			expectedOutput: []string{
				`apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'statefulset_deployment-foo.yaml'
`,
				`apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
`,
			},
			expectedError: "",
			instance: exec.Filter{
				Path: "sed",
				Args: []string{"s/Deployment/StatefulSet/g"},
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			// initialize the inputs for the FunctionFilter
			var inputs []*yaml.RNode
			for i := range tt.input {
				node, err := yaml.Parse(tt.input[i])
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				inputs = append(inputs, node)
			}
			if tt.functionConfig != "" {
				fc, err := yaml.Parse(tt.functionConfig)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				tt.instance.FunctionConfig = fc
			}

			// run the function
			output, err := tt.instance.Filter(inputs)

			// check for errors
			if tt.expectedError != "" {
				if !assert.EqualError(t, err, tt.expectedError) {
					t.FailNow()
				}
				return
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// verify the output
			var actual []string
			for i := range output {
				s, err := output[i].String()
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				actual = append(actual, strings.TrimSpace(s))
			}
			var expected []string
			for i := range tt.expectedOutput {
				expected = append(expected, strings.TrimSpace(tt.expectedOutput[i]))
			}
			if !assert.Equal(t, expected, actual) {
				t.FailNow()
			}
		})
	}
}
