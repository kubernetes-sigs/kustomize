// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package runtimeutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type testRun struct {
	err           error
	expectedInput string
	output        string
	t             *testing.T
}

func (r testRun) run(reader io.Reader, writer io.Writer) error {
	if r.expectedInput != "" {
		input, err := ioutil.ReadAll(reader)
		if !assert.NoError(r.t, err) {
			r.t.FailNow()
		}

		// verify input matches expected
		if !assert.Equal(r.t, r.expectedInput, string(input)) {
			r.t.FailNow()
		}
	}

	_, err := writer.Write([]byte(r.output))
	if !assert.NoError(r.t, err) {
		r.t.FailNow()
	}

	return r.err
}

func TestFunctionFilter_Filter(t *testing.T) {
	var tests = []struct {
		run                testRun
		name               string
		input              []string
		functionConfig     string
		expectedOutput     []string
		expectedError      string
		expectedSavedError string
		expectedResults    string
		noMakeResultsFile  bool
		instance           FunctionFilter
	}{
		// verify that resources emitted from the function have a file path defaulted
		// if none already exists
		{
			name: "default_file_path_annotation",
			run: testRun{
				output: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
`,
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
`,
			},
		},

		// verify that resources emitted from the function do not have a file path defaulted
		// if one already exists
		{
			name: "no_default_file_path_annotation",
			run: testRun{
				output: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
     config.kubernetes.io/path: 'foo.yaml'
`,
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo.yaml'
`,
			},
		},

		// verify the FunctionFilter correctly writes the inputs and reads the outputs
		// of Run
		{
			name: "write_read",
			run: testRun{
				output: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo.yaml'
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: configmap-foo
    annotations:
      config.kubernetes.io/path: 'foo.yaml'
`,
			},
			input: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`,
			},
			expectedOutput: []string{`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo.yaml'
`,
				`
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap-foo
  annotations:
    config.kubernetes.io/path: 'foo.yaml'
`,
			},
		},

		// verify that the results file is written
		//
		{
			name: "write_results_file",
			run: testRun{
				output: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
results:
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
`,
			},
			expectedOutput: []string{`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
`, `
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
`,
			},
			expectedResults: `
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
`,
		},

		// verify that the results file is written for functions that exist non-0
		// and the FunctionFilter returns the error
		{
			name:          "write_results_file_function_exit_non_0",
			expectedError: "failed",
			run: testRun{
				err: fmt.Errorf("failed"),
				output: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
results:
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
`,
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
    `, `
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
    `,
			},
			expectedResults: `
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
    `,
		},

		// verify that if deferFailure is set, the results file is written and the
		// exit error is saved, but the FunctionFilter does not return an error.
		{
			name:               "write_results_defer_failure",
			instance:           FunctionFilter{DeferFailure: true},
			expectedSavedError: "failed",
			run: testRun{
				err: fmt.Errorf("failed"),
				output: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
results:
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"`,
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
		`, `
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
		`,
			},
			expectedResults: `
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"`,
		},

		{
			name:              "write_results_bad_results_file",
			expectedError:     "open /not/real/file:",
			noMakeResultsFile: true,
			run: testRun{
				output: `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
results:
- apiVersion: config.k8s.io/v1alpha1
  name: "some-validator"
`,
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
		`, `
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
		`,
			},
			// these aren't written, expect an error
			expectedResults: `
- apiVersion: config.k8s.io/v1alpha1
  name: "some-validator"
`,
		},

		// verify the function only sees resources scoped to it based on the directory
		// containing the functionConfig and the directory containing each resource.
		// resources not provided to the function should still appear in the FunctionFilter
		// output
		{
			name: "scope_resources_by_directory",
			run: testRun{
				expectedInput: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.k8s.io/id: '1'
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/bar.yaml'
`,
				output: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      new: annotation
      config.k8s.io/id: '1'
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/bar.yaml'
`,
			},
			functionConfig: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`,
			input: []string{
				// this should not be in scope
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'baz/bar/d.yaml'
`,
				// this should be in scope
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`},
			expectedOutput: []string{
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    new: annotation
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'baz/bar/d.yaml'
`,
			},
		},

		// verify functions without file path annotation are not scoped to functions
		{
			name: "scope_resources_by_directory_resources_missing_path",
			run: testRun{
				expectedInput: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.k8s.io/id: '1'
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/bar.yaml'
`,
				output: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      new: annotation
      config.k8s.io/id: '1'
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/bar.yaml'
`,
			},
			functionConfig: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`,
			input: []string{
				// this should not be in scope
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
`,
				// this should be in scope
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`},
			expectedOutput: []string{
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    new: annotation
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
`,
			},
		},

		// verify the functions can see all resources if global scope is set
		{
			name:     "scope_resources_global",
			instance: FunctionFilter{GlobalScope: true},
			run: testRun{
				expectedInput: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'baz/bar/d.yaml'
      config.k8s.io/id: '1'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.k8s.io/id: '2'
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/bar.yaml'
`,
				output: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'baz/bar/d.yaml'
      config.k8s.io/id: '1'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      new: annotation
      config.k8s.io/id: '2'
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/bar.yaml'
`,
			},
			functionConfig: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`,
			input: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'baz/bar/d.yaml'
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'baz/bar/d.yaml'
`, `
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    new: annotation
`,
			},
		},

		{
			name: "scope_no_resources",
			run: testRun{
				expectedInput: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/bar.yaml'
`,
				output: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/bar.yaml'
`,
			},
			functionConfig: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`,
			input: []string{
				// these should not be in scope
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'baz/bar/d.yaml'
`, `
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'biz/bar/s.yaml'
`},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'baz/bar/d.yaml'
`, `
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'biz/bar/s.yaml'
`,
			},
		},

		{
			name: "scope_functions_dir",
			run: testRun{
				expectedInput: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.k8s.io/id: '1'
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/functions/bar.yaml'
`,
				output: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.k8s.io/id: '1'
      new: annotation
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/functions/bar.yaml'
`,
			},
			functionConfig: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/functions/bar.yaml'
`,
			input: []string{
				// this should not be in scope
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'baz/bar/d.yaml'
`,
				// this should be in scope
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`},
			expectedOutput: []string{
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    new: annotation
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'baz/bar/d.yaml'
`,
			},
		},

		{
			name: "copy_comments",
			run: testRun{
				expectedInput: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'foo/b.yaml'
      config.k8s.io/id: '1'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo # name comment
    annotations:
      config.kubernetes.io/path: 'foo/a.yaml'
      config.k8s.io/id: '2'
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/f.yaml'
`,
				// delete the comment
				output: `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'foo/b.yaml'
      config.k8s.io/id: '1'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/a.yaml'
      config.k8s.io/id: '2'
      new: annotation
functionConfig:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/path: 'foo/f.yaml'
`,
			},
			functionConfig: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/f.yaml'
`,
			input: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/b.yaml'
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo # name comment
  annotations:
    config.kubernetes.io/path: 'foo/a.yaml'
`},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/b.yaml'
`, `
apiVersion: v1
kind: Service
metadata:
  name: service-foo # name comment
  annotations:
    config.kubernetes.io/path: 'foo/a.yaml'
    new: annotation
`,
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			// results file setup
			if len(tt.expectedResults) > 0 && !tt.noMakeResultsFile {
				// expect result files to be written -- create a directory for them
				f, err := ioutil.TempFile("", "test-kyaml-*.yaml")
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				defer os.RemoveAll(f.Name())
				tt.instance.ResultsFile = f.Name()
			} else if len(tt.expectedResults) > 0 {
				// failure case for writing to bad results location
				tt.instance.ResultsFile = "/not/real/file"
			}

			// initialize the inputs for the FunctionFilter
			var inputs []*yaml.RNode
			for i := range tt.input {
				node, err := yaml.Parse(tt.input[i])
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				inputs = append(inputs, node)
			}

			// run the FunctionFilter
			tt.run.t = t
			tt.instance.Run = tt.run.run
			if tt.functionConfig != "" {
				fc, err := yaml.Parse(tt.functionConfig)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				tt.instance.FunctionConfig = fc
			}
			output, err := tt.instance.Filter(inputs)
			if tt.expectedError != "" {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), tt.expectedError) {
					t.FailNow()
				}
				return
			}

			// check for saved error
			if tt.expectedSavedError != "" {
				if !assert.EqualError(t, tt.instance.exit, tt.expectedSavedError) {
					t.FailNow()
				}
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// verify function output
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

			// verify results files
			if len(tt.instance.ResultsFile) > 0 {
				tt.expectedResults = strings.TrimSpace(tt.expectedResults)

				results, err := tt.instance.results.String()
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tt.expectedResults, strings.TrimSpace(results)) {
					t.FailNow()
				}

				b, err := ioutil.ReadFile(tt.instance.ResultsFile)
				writtenResults := strings.TrimSpace(string(b))
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tt.expectedResults, writtenResults) {
					t.FailNow()
				}
			}
		})
	}
}

func Test_GetFunction(t *testing.T) {
	var tests = []struct {
		name       string
		resource   string
		expectedFn string
		missingFn  bool
	}{

		// fn annotation
		{
			name: "fn annotation",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
`,
			expectedFn: `
container:
    image: foo:v1.0.0`,
		},

		{
			name: "storage mounts json style",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
        mounts: [ {type: bind, src: /mount/path, dst: /local/}, {src: myvol, dst: /local/, type: volume}, {dst: /local/, type: tmpfs} ]
`,
			expectedFn: `
container:
    image: foo:v1.0.0
    mounts:
      - type: bind
        src: /mount/path
        dst: /local/
      - type: volume
        src: myvol
        dst: /local/
      - type: tmpfs
        dst: /local/
`,
		},

		{
			name: "storage mounts yaml style",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
        mounts:
        - src: /mount/path
          type: bind
          dst: /local/
        - dst: /local/
          src: myvol
          type: volume
        - type: tmpfs
          dst: /local/
`,
			expectedFn: `
container:
    image: foo:v1.0.0
    mounts:
      - type: bind
        src: /mount/path
        dst: /local/
      - type: volume
        src: myvol
        dst: /local/
      - type: tmpfs
        dst: /local/
`,
		},

		{
			name: "network",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
        network: true
`,
			expectedFn: `
container:
    image: foo:v1.0.0
    network: true
`,
		},

		{
			name: "path",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      path: foo
      container:
        image: foo:v1.0.0
`,
			// path should be erased
			expectedFn: `
container:
    image: foo:v1.0.0
`,
		},

		{
			name: "network",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      network: foo
      container:
        image: foo:v1.0.0
`,
			// network should be erased
			expectedFn: `
container:
    image: foo:v1.0.0
`,
		},

		// legacy fn style
		{name: "legacy fn meta",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  configFn:
      container:
        image: foo:v1.0.0
`,
			expectedFn: `
container:
    image: foo:v1.0.0
`,
		},

		// no fn
		{name: "no fn",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations: {}
`,
			missingFn: true,
		},

		// test network, etc...
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			resource := yaml.MustParse(tt.resource)
			fn := GetFunctionSpec(resource)
			if tt.missingFn {
				if !assert.Nil(t, fn) {
					t.FailNow()
				}
			} else {
				b, err := yaml.Marshal(fn)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(tt.expectedFn),
					strings.TrimSpace(string(b))) {
					t.FailNow()
				}
			}
		})
	}
}

func Test_GetContainerNetworkRequired(t *testing.T) {
	tests := []struct {
		input    string
		required bool
	}{
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
      network: true
`,
			required: true,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
      network: false
`,
			required: false,
		},
		{

			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
`,
			required: false,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
        network: true
`,
			required: true,
		},
	}

	for _, tc := range tests {
		cfg, err := yaml.Parse(tc.input)
		if !assert.NoError(t, err) {
			return
		}
		fn := GetFunctionSpec(cfg)
		assert.Equal(t, tc.required, fn.Container.Network)
	}
}

func Test_StringToStorageMount(t *testing.T) {
	tests := []struct {
		in          string
		expectedOut string
	}{
		{
			in:          "type=bind,src=/tmp/test/,dst=/tmp/source/",
			expectedOut: "type=bind,source=/tmp/test/,target=/tmp/source/,readonly",
		},
		{
			in:          "type=bind,src=/tmp/test/,dst=/tmp/source/,rw=true",
			expectedOut: "type=bind,source=/tmp/test/,target=/tmp/source/",
		},
		{
			in:          "type=bind,src=/tmp/test/,dst=/tmp/source/,rw=false",
			expectedOut: "type=bind,source=/tmp/test/,target=/tmp/source/,readonly",
		},
		{
			in:          "type=bind,src=/tmp/test/,dst=/tmp/source/,rw=",
			expectedOut: "type=bind,source=/tmp/test/,target=/tmp/source/,readonly",
		},
		{
			in:          "type=tmpfs,src=/tmp/test/,dst=/tmp/source/,rw=invalid",
			expectedOut: "type=tmpfs,source=/tmp/test/,target=/tmp/source/,readonly",
		},
		{
			in:          "type=tmpfs,src=/tmp/test/,dst=/tmp/source/,rwe=invalid",
			expectedOut: "type=tmpfs,source=/tmp/test/,target=/tmp/source/,readonly",
		},
		{
			in:          "type=tmpfs,src=/tmp/test/,dst",
			expectedOut: "type=tmpfs,source=/tmp/test/,target=,readonly",
		},
		{
			in:          "type=bind,source=/tmp/test/,target=/tmp/source/,rw=true",
			expectedOut: "type=bind,source=/tmp/test/,target=/tmp/source/",
		},
		{
			in:          "type=bind,source=/tmp/test/,target=/tmp/source/",
			expectedOut: "type=bind,source=/tmp/test/,target=/tmp/source/,readonly",
		},
	}

	for _, tc := range tests {
		s := StringToStorageMount(tc.in)
		assert.Equal(t, tc.expectedOut, (&s).String())
	}
}

func TestContainerEnvGetDockerFlags(t *testing.T) {
	tests := []struct {
		input  *ContainerEnv
		output []string
	}{
		{
			input:  NewContainerEnvFromStringSlice([]string{"foo=bar"}),
			output: []string{"-e", "LOG_TO_STDERR=true", "-e", "STRUCTURED_RESULTS=true", "-e", "foo=bar"},
		},
		{
			input:  NewContainerEnvFromStringSlice([]string{"foo"}),
			output: []string{"-e", "LOG_TO_STDERR=true", "-e", "STRUCTURED_RESULTS=true", "-e", "foo"},
		},
		{
			input:  NewContainerEnvFromStringSlice([]string{"foo=bar", "baz"}),
			output: []string{"-e", "LOG_TO_STDERR=true", "-e", "STRUCTURED_RESULTS=true", "-e", "foo=bar", "-e", "baz"},
		},
		{
			input:  NewContainerEnv(),
			output: []string{"-e", "LOG_TO_STDERR=true", "-e", "STRUCTURED_RESULTS=true"},
		},
	}

	for _, tc := range tests {
		flags := tc.input.GetDockerFlags()
		assert.Equal(t, tc.output, flags)
	}
}

func TestGetContainerEnv(t *testing.T) {
	tests := []struct {
		input    string
		expected ContainerEnv
	}{
		{
			input: `
apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
      envs:
      - foo=bar
`,
			expected: *NewContainerEnvFromStringSlice([]string{"foo=bar"}),
		},
		{
			input: `
apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
      envs:
      - foo=bar
      - baz
`,
			expected: *NewContainerEnvFromStringSlice([]string{"foo=bar", "baz"}),
		},
		{
			input: `
apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
      envs:
      - KUBECONFIG
`,
			expected: *NewContainerEnvFromStringSlice([]string{"KUBECONFIG"}),
		},
	}

	for _, tc := range tests {
		cfg, err := yaml.Parse(tc.input)
		if !assert.NoError(t, err) {
			return
		}
		fn := GetFunctionSpec(cfg)
		assert.Equal(t, tc.expected, *NewContainerEnvFromStringSlice(fn.Container.Env))
	}
}
