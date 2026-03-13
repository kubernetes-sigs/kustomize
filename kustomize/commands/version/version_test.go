// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionOutput(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "default output succeeds",
			output:    "",
			expectErr: false,
		},
		{
			name:      "yaml output succeeds",
			output:    "yaml",
			expectErr: false,
		},
		{
			name:      "json output succeeds",
			output:    "json",
			expectErr: false,
		},
		{
			name:      "invalid output yml returns error",
			output:    "yml",
			expectErr: true,
			errMsg:    "--output must be 'yaml' or 'json'",
		},
		{
			name:      "invalid output text returns error",
			output:    "text",
			expectErr: true,
			errMsg:    "--output must be 'yaml' or 'json'",
		},
		{
			name:      "invalid output arbitrary string returns error",
			output:    "foo",
			expectErr: true,
			errMsg:    "--output must be 'yaml' or 'json'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			o := &Options{
				Output: tt.output,
				Writer: &buf,
			}
			err := o.Run()
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}

func TestVersionValidate(t *testing.T) {
	tests := []struct {
		name      string
		short     bool
		output    string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "no flags",
			expectErr: false,
		},
		{
			name:      "short only",
			short:     true,
			expectErr: false,
		},
		{
			name:      "output only",
			output:    "json",
			expectErr: false,
		},
		{
			name:      "short and output are mutually exclusive",
			short:     true,
			output:    "json",
			expectErr: true,
			errMsg:    "--short and --output are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				Short:  tt.short,
				Output: tt.output,
			}
			err := o.Validate(nil)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
