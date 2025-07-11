// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShlexSplit(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name:     "basic space separation",
			input:    `hello world`,
			expected: []string{"hello", "world"},
			wantErr:  false,
		},
		{
			name:     "double quoted string",
			input:    `"hello world"`,
			expected: []string{"hello world"},
			wantErr:  false,
		},
		{
			name:     "single quoted string",
			input:    `'hello world'`,
			expected: []string{"hello world"},
			wantErr:  false,
		},
		{
			name:     "mixed quotes and words",
			input:    `hello "world test"`,
			expected: []string{"hello", "world test"},
			wantErr:  false,
		},
		{
			name:     "single quotes with spaces",
			input:    `hello 'world test'`,
			expected: []string{"hello", "world test"},
			wantErr:  false,
		},
		{
			name:     "nested quotes - single in double",
			input:    `"hello 'nested' world"`,
			expected: []string{"hello 'nested' world"},
			wantErr:  false,
		},
		{
			name:     "nested quotes - double in single",
			input:    `'hello "nested" world'`,
			expected: []string{"hello \"nested\" world"},
			wantErr:  false,
		},
		{
			name:     "escaped space",
			input:    `hello\ world`,
			expected: []string{"hello world"},
			wantErr:  false,
		},
		{
			name:     "escaped quotes in double quotes",
			input:    `"hello \"world\""`,
			expected: []string{"hello \"world\""},
			wantErr:  false,
		},
		{
			name:     "single quote in single quotes",
			input:    `'can'\''t'`,
			expected: []string{"can't"},
			wantErr:  false,
		},
		{
			name:     "complex argument list",
			input:    `arg1 "arg 2" 'arg 3' arg4`,
			expected: []string{"arg1", "arg 2", "arg 3", "arg4"},
			wantErr:  false,
		},
		{
			name:     "echo command with escaped quotes",
			input:    `echo "Hello, \"World!\""`,
			expected: []string{"echo", "Hello, \"World!\""},
			wantErr:  false,
		},
		{
			name:     "grep command with quoted search term",
			input:    `grep -r "search term" /path/to/dir`,
			expected: []string{"grep", "-r", "search term", "/path/to/dir"},
			wantErr:  false,
		},
		{
			name:     "ls command with quoted filename",
			input:    `ls -la "file with spaces.txt"`,
			expected: []string{"ls", "-la", "file with spaces.txt"},
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    ``,
			expected: []string{""},
			wantErr:  false,
		},
		{
			name:     "multiple spaces",
			input:    `   multiple   spaces   `,
			expected: []string{"multiple", "spaces", ""},
			wantErr:  false,
		},
		{
			name:     "with comment string",
			input:    `echo "Hello, W#orld!" ${USER} # This is a comment`,
			expected: []string{"echo", "Hello, W#orld!", "${USER}"},
			wantErr:  false,
		},
		// may cause an error in shlex at python3
		{
			name:     "unclosed double quote",
			input:    `"unclosed quote`,
			expected: []string{"unclosed quote"},
			wantErr:  false,
		},
		{
			name:     "unclosed single quote",
			input:    `'unclosed quote`,
			expected: []string{"unclosed quote"},
			wantErr:  false,
		},
		{
			name:     "mixed unclosed quotes",
			input:    `"mixed 'quotes`,
			expected: []string{"mixed 'quotes"},
			wantErr:  false,
		},
		{
			name:     "single quote closed with double quote",
			input:    `"hello world'`,
			expected: []string{"hello world'"},
			wantErr:  false,
		},
		{
			name:     "double quote closed with single quote",
			input:    `'hello world"`,
			expected: []string{"hello world\""},
			wantErr:  false,
		},
	}

	// execute each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// call the ShlexSplit function
			result, err := ShlexSplit(tc.input)

			// check for expected error
			if tc.wantErr {
				if err == nil {
					t.Errorf("FAIL: Expected error but got none, Expected: %q\n", tc.expected)
				}
				return
			}

			if assert.NoError(t, err, "FAIL: Unexpected error for input %q", tc.input) {
				// check if the result matches the expected output
				assert.Equal(t, tc.expected, result,
					"FAIL: Result mismatch,Input %q, Expected %q, Got: %q\n",
					tc.input, tc.expected, result,
				)
			}
		})
	}
}
