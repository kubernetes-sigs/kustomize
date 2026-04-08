// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newOptionsWithOutput returns an Options with the given output format and a
// bytes.Buffer as its writer, preventing potential nil-Writer panics in Run().
func newOptionsWithOutput(output string) *Options {
	o := NewOptions(&bytes.Buffer{})
	o.Output = output
	return o
}

// newOptionsWithShortAndOutput returns an Options with the given short and
// output settings and a bytes.Buffer as its writer.
func newOptionsWithShortAndOutput(short bool, output string) *Options {
	o := NewOptions(&bytes.Buffer{})
	o.Short = short
	o.Output = output
	return o
}

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *Options
		wantErr string
	}{
		{
			name:    "valid empty output",
			opts:    NewOptions(&bytes.Buffer{}),
			wantErr: "",
		},
		{
			name:    "valid yaml output",
			opts:    newOptionsWithOutput("yaml"),
			wantErr: "",
		},
		{
			name:    "valid json output",
			opts:    newOptionsWithOutput("json"),
			wantErr: "",
		},
		{
			name:    "invalid output yml",
			opts:    newOptionsWithOutput("yml"),
			wantErr: "--output must be 'yaml' or 'json'",
		},
		{
			name:    "invalid output xml",
			opts:    newOptionsWithOutput("xml"),
			wantErr: "--output must be 'yaml' or 'json'",
		},
		{
			name:    "short and output mutually exclusive",
			opts:    newOptionsWithShortAndOutput(true, "yaml"),
			wantErr: "--short and --output are mutually exclusive",
		},
		{
			name:    "short fires before invalid output",
			opts:    newOptionsWithShortAndOutput(true, "yml"),
			wantErr: "--short and --output are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate(nil)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestOptions_Run_ValidOutputs(t *testing.T) {
	tests := []struct {
		name   string
		output string
		check  func(t *testing.T, buf *bytes.Buffer)
	}{
		{
			name:   "yaml output",
			output: "yaml",
			check: func(t *testing.T, buf *bytes.Buffer) {
				assert.Greater(t, buf.Len(), 0, "expected non-empty yaml output")
				assert.Contains(t, buf.String(), "version:")
			},
		},
		{
			name:   "json output",
			output: "json",
			check: func(t *testing.T, buf *bytes.Buffer) {
				assert.Greater(t, buf.Len(), 0, "expected non-empty json output")
				assert.True(t, strings.HasPrefix(buf.String(), "{"), "expected json output to start with '{'")
				assert.Contains(t, buf.String(), "version")
			},
		},
		{
			name:   "default output",
			output: "",
			check: func(t *testing.T, buf *bytes.Buffer) {
				assert.Greater(t, buf.Len(), 0, "expected non-empty default output")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			o := NewOptions(buf)
			o.Output = tt.output
			err := o.Run()
			require.NoError(t, err)
			tt.check(t, buf)
		})
	}
}

func TestOptions_Run_InvalidOutputReturnsError(t *testing.T) {
	o := newOptionsWithOutput("yml")
	err := o.Run()
	require.EqualError(t, err, "--output must be 'yaml' or 'json'")
}

func TestNewCmdVersion_InvalidOutputFlag(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := NewCmdVersion(buf)
	cmd.SilenceErrors = true // prevent cobra from printing to os.Stderr
	cmd.SilenceUsage = true  // prevent cobra from printing usage to os.Stderr
	cmd.SetArgs([]string{"--output=yml"})
	err := cmd.Execute()
	require.EqualError(t, err, "--output must be 'yaml' or 'json'")
}
