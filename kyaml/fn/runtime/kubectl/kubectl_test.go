// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubectl

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
)

func Test_getPodConfig(t *testing.T) {
	var tests = []struct {
		ContainerSpec     runtimeutil.ContainerSpec
		PodTemplateName   string
		KubectlGlobalArgs []string
		PodStartTimeout   *time.Duration
		ExpectedErr       bool
		ExpectedPodConfig string
	}{
		{
			ContainerSpec: runtimeutil.ContainerSpec{
				Image: "example.com:version",
			},
			ExpectedPodConfig: `
apiVersion: v1
kind: Pod
metadata:
  name: krm-example
  labels:
    app: krm-pod
spec:
  containers:
  - name: default
    stdin: true
    stdinOnce: true
    image: example.com:version
    env: [{"name": "LOG_TO_STDERR", "value": "true"}, {"name": "STRUCTURED_RESULTS", "value": "true"}]
  restartPolicy: Never
`,
		},
		{
			ContainerSpec: runtimeutil.ContainerSpec{
				Image: "example.com:version",
				Env: []string{
					"a=a",
					"b",
				},
			},
			ExpectedPodConfig: `
apiVersion: v1
kind: Pod
metadata:
  name: krm-example
  labels:
    app: krm-pod
spec:
  containers:
  - name: default
    stdin: true
    stdinOnce: true
    image: example.com:version
    env: [{"name": "LOG_TO_STDERR", "value": "true"}, {"name": "STRUCTURED_RESULTS", "value": "true"}, {"name": "a", "value": "a"}, {"name": "b", "value": ""}]
  restartPolicy: Never
`,
		},
		{
			PodTemplateName: "doesntExist",
			ExpectedErr:     true,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			o := NewContainer(tt.ContainerSpec, tt.PodTemplateName, tt.KubectlGlobalArgs, tt.PodStartTimeout)

			podConfig, err := o.getPodConfig("krm-example")
			if tt.ExpectedErr {
				if err == nil {
					t.FailNow()
				}
			} else {
				if err != nil {
					t.FailNow()
				}
				if !assert.Equal(t, tt.ExpectedPodConfig[1:], podConfig) {
					t.FailNow()
				}
			}
		})
	}
}

func Test_mergeYamlStrings(t *testing.T) {
	var tests = []struct {
		Y1          string
		Y2          string
		Expected    string
		ExpectedErr bool
	}{
		{
			Y1: `[{"name": "LOG_TO_STDERR", "value": "true"}, {"name": "STRUCTURED_RESULTS", "value": "true"}, ]`,
			Y2: `
- name: x1
  value: y1
- name: x2
  value: y2
`,
			Expected: `
- name: LOG_TO_STDERR
  value: "true"
- name: STRUCTURED_RESULTS
  value: "true"
- name: x1
  value: y1
- name: x2
  value: y2
`,
		},
	}
	for i := range tests {
		r, _ := mergeYamlStrings(tests[i].Y1, tests[i].Y2)
		if !assert.Equal(t, tests[i].Expected[1:], r) {
			t.FailNow()
		}
	}
}
