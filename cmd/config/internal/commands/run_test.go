// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestRunFnCommand_preRunE verifies that preRunE correctly parses the commandline
// flags and arguments into the RunFns structure to be executed.
func TestRunFnCommand_preRunE(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expected      string
		err           string
		path          string
		input         io.Reader
		output        io.Writer
		functionPaths []string
	}{
		{
			name: "config map",
			args: []string{"run", "dir", "--image", "foo:bar", "--", "a=b", "c=d", "e=f"},
			path: "dir",
			expected: `
metadata:
  name: function-input
  annotations:
    config.kubernetes.io/function: |
      container: {image: 'foo:bar'}
data: {a: b, c: d, e: f}
kind: ConfigMap
apiVersion: v1
`,
		},
		{
			name:   "config map stdin / stdout",
			args:   []string{"run", "--image", "foo:bar", "--", "a=b", "c=d", "e=f"},
			input:  os.Stdin,
			output: os.Stdout,
			expected: `
metadata:
  name: function-input
  annotations:
    config.kubernetes.io/function: |
      container: {image: 'foo:bar'}
data: {a: b, c: d, e: f}
kind: ConfigMap
apiVersion: v1
`,
		},
		{
			name:   "config map dry-run",
			args:   []string{"run", "dir", "--image", "foo:bar", "--dry-run", "--", "a=b", "c=d", "e=f"},
			output: os.Stdout,
			path:   "dir",
			expected: `
metadata:
  name: function-input
  annotations:
    config.kubernetes.io/function: |
      container: {image: 'foo:bar'}
data: {a: b, c: d, e: f}
kind: ConfigMap
apiVersion: v1
`,
		},
		{
			name: "config map no args",
			args: []string{"run", "dir", "--image", "foo:bar"},
			path: "dir",
			expected: `
metadata:
  name: function-input
  annotations:
    config.kubernetes.io/function: |
      container: {image: 'foo:bar'}
data: {}
kind: ConfigMap
apiVersion: v1
`,
		},
		{
			name: "custom kind",
			args: []string{"run", "dir", "--image", "foo:bar", "--", "Foo", "g=h"},
			path: "dir",
			expected: `
metadata:
  name: function-input
  annotations:
    config.kubernetes.io/function: |
      container: {image: 'foo:bar'}
data: {g: h}
kind: Foo
apiVersion: v1
`,
		},
		{
			name: "custom kind '=' in data",
			args: []string{"run", "dir", "--image", "foo:bar", "--", "Foo", "g=h", "i=j=k"},
			path: "dir",
			expected: `
metadata:
  name: function-input
  annotations:
    config.kubernetes.io/function: |
      container: {image: 'foo:bar'}
data: {g: h, i: j=k}
kind: Foo
apiVersion: v1
`,
		},
		{
			name:          "function paths",
			args:          []string{"run", "dir", "--fn-path", "path1", "--fn-path", "path2"},
			path:          "dir",
			functionPaths: []string{"path1", "path2"},
		},
		{
			name: "custom kind with function paths",
			args: []string{
				"run", "dir", "--fn-path", "path", "--image", "foo:bar", "--", "Foo", "g=h", "i=j=k"},
			path:          "dir",
			functionPaths: []string{"path"},
			expected: `
metadata:
  name: function-input
  annotations:
    config.kubernetes.io/function: |
      container: {image: 'foo:bar'}
data: {g: h, i: j=k}
kind: Foo
apiVersion: v1
`,
		},
		{
			name: "config map multi args",
			args: []string{"run", "dir", "dir2", "--image", "foo:bar", "--", "a=b", "c=d", "e=f"},
			err:  "0 or 1 arguments supported",
		},
		{
			name: "config map not image",
			args: []string{"run", "dir", "--", "a=b", "c=d", "e=f"},
			err:  "must specify --image",
		},
		{
			name: "config map bad data",
			args: []string{"run", "dir", "--image", "foo:bar", "--", "a=b", "c", "e=f"},
			err:  "must have keys and values separated by",
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			r := GetRunFnRunner("kustomize")
			// Don't run the actual command
			r.Command.Run = nil
			r.Command.RunE = func(cmd *cobra.Command, args []string) error { return nil }
			r.Command.SilenceErrors = true
			r.Command.SilenceUsage = true

			// hack due to https://github.com/spf13/cobra/issues/42
			root := &cobra.Command{Use: "root"}
			root.AddCommand(r.Command)
			root.SetArgs(tt.args)

			// error case
			err := r.Command.Execute()
			if tt.err != "" {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), tt.err) {
					t.FailNow()
				}
				// don't check anything else in error case
				return
			}

			// non-error case
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// check if Input was set
			if !assert.Equal(t, tt.input, r.RunFns.Input) {
				t.FailNow()
			}

			// check if Output was set
			if !assert.Equal(t, tt.output, r.RunFns.Output) {
				t.FailNow()
			}

			// check if Path was set
			if !assert.Equal(t, tt.path, r.RunFns.Path) {
				t.FailNow()
			}

			// check if FunctionPaths were set
			if tt.functionPaths == nil {
				// make Equal work against flag default
				tt.functionPaths = []string{}
			}
			if !assert.Equal(t, tt.functionPaths, r.RunFns.FunctionPaths) {
				t.FailNow()
			}

			// check if Functions were set
			if tt.expected != "" {
				if !assert.Len(t, r.RunFns.Functions, 1) {
					t.FailNow()
				}
				actual := strings.TrimSpace(r.RunFns.Functions[0].MustString())
				if !assert.Equal(t, strings.TrimSpace(tt.expected), actual) {
					t.FailNow()
				}
			}

		})
	}

}
