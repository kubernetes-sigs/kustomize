// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionValidation(t *testing.T) {
	testCases := []struct {
		name        string
		output      string
		short       bool
		expectedErr string
	}{
		{
			name:   "default output",
			output: "",
		},
		{
			name:   "yaml output",
			output: "yaml",
		},
		{
			name:   "json output",
			output: "json",
		},
		{
			name:        "invalid output yml",
			output:      "yml",
			expectedErr: "--output must be 'yaml' or 'json'",
		},
		{
			name:        "invalid output text",
			output:      "text",
			expectedErr: "--output must be 'yaml' or 'json'",
		},
		{
			name:        "short and output mutually exclusive",
			output:      "yaml",
			short:       true,
			expectedErr: "--short and --output are mutually exclusive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o := &Options{
				Output: tc.output,
				Short:  tc.short,
				Writer: &bytes.Buffer{},
			}
			err := o.Validate(nil)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVersionRun(t *testing.T) {
	testCases := []struct {
		name       string
		output     string
		short      bool
		wantPrefix string
	}{
		{
			name:   "default format",
			output: "",
		},
		{
			name:       "short format",
			short:      true,
			wantPrefix: "{",
		},
		{
			name:   "yaml format",
			output: "yaml",
		},
		{
			name:       "json format",
			output:     "json",
			wantPrefix: "{\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			o := &Options{
				Output: tc.output,
				Short:  tc.short,
				Writer: buf,
			}
			require.NoError(t, o.Validate(nil))
			require.NoError(t, o.Run())
			assert.NotEmpty(t, buf.String())
			if tc.wantPrefix != "" {
				assert.True(t, bytes.HasPrefix(buf.Bytes(), []byte(tc.wantPrefix)))
			}
		})
	}
}

func TestNewCmdVersion_InvalidOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := NewCmdVersion(buf)
	cmd.SetArgs([]string{"--output", "invalid"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "--output must be 'yaml' or 'json'")
}
