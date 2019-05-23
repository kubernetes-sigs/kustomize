/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package expansion_test

import (
	"fmt"
	"testing"

	. "sigs.k8s.io/kustomize/pkg/expansion"
)

func TestInlineMapReference(t *testing.T) {
	type env struct {
		Name  string
		Value interface{}
	}
	envs := []env{
		{
			Name:  "FOO",
			Value: "bar",
		},
		{
			Name:  "ZOO",
			Value: "$(FOO)",
		},
	}

	declaredEnv := map[string]interface{}{
		"FOO": "bar",
		"ZOO": "$(FOO)",
	}

	counts := make(map[string]int)
	mapping := InlineFuncFor(counts, declaredEnv)

	for _, env := range envs {
		declaredEnv[env.Name] = Inline(fmt.Sprintf("%v", env.Value), mapping)
	}

	expectedEnv := map[string]expected{
		"FOO": {count: 1, edited: "bar"},
		"ZOO": {count: 0, edited: "bar"},
	}

	for k, v := range expectedEnv {
		if e, a := v, declaredEnv[k]; e.edited != a || e.count != counts[k] {
			t.Errorf("Expected %v count=%d, got %v count=%d",
				e.edited, e.count, a, counts[k])
		} else {
			delete(declaredEnv, k)
		}
	}

	if len(declaredEnv) != 0 {
		t.Errorf("Unexpected keys in declared env: %v", declaredEnv)
	}
}

func TestInlineMapping(t *testing.T) {
	context := map[string]interface{}{
		"VAR_A":     "A",
		"VAR_B":     "B",
		"VAR_C":     "C",
		"VAR_REF":   "$(VAR_A)",
		"VAR_EMPTY": "",
	}
	doInlineTest(t, context)
}

func TestInlineMappingDual(t *testing.T) {
	context := map[string]interface{}{
		"VAR_A":     "A",
		"VAR_EMPTY": "",
	}
	context2 := map[string]interface{}{
		"VAR_B":   "B",
		"VAR_C":   "C",
		"VAR_REF": "$(VAR_A)",
	}

	doInlineTest(t, context, context2)
}

func doInlineTest(t *testing.T, context ...map[string]interface{}) {
	cases := []struct {
		name     string
		input    string
		expected string
		counts   map[string]int
	}{
		{
			name:     "whole string",
			input:    "$(VAR_A)",
			expected: "A",
			counts:   map[string]int{"VAR_A": 1},
		},
		{
			name:     "repeat",
			input:    "$(VAR_A)-$(VAR_A)",
			expected: "$(VAR_A)-$(VAR_A)",
		},
		{
			name:     "multiple repeats",
			input:    "$(VAR_A)-$(VAR_B)-$(VAR_B)-$(VAR_B)-$(VAR_A)",
			expected: "$(VAR_A)-$(VAR_B)-$(VAR_B)-$(VAR_B)-$(VAR_A)",
		},
		{
			name:     "beginning",
			input:    "$(VAR_A)-1",
			expected: "$(VAR_A)-1",
		},
		{
			name:     "middle",
			input:    "___$(VAR_B)___",
			expected: "___$(VAR_B)___",
		},
		{
			name:     "end",
			input:    "___$(VAR_C)",
			expected: "___$(VAR_C)",
		},
		{
			name:     "compound",
			input:    "$(VAR_A)_$(VAR_B)_$(VAR_C)",
			expected: "$(VAR_A)_$(VAR_B)_$(VAR_C)",
		},
		{
			name:     "escape & expand",
			input:    "$$(VAR_B)_$(VAR_A)",
			expected: "$$(VAR_B)_$(VAR_A)",
		},
		{
			name:     "compound escape",
			input:    "$$(VAR_A)_$$(VAR_B)",
			expected: "$$(VAR_A)_$$(VAR_B)",
		},
		{
			name:     "mixed in escapes",
			input:    "f000-$$VAR_A",
			expected: "f000-$$VAR_A",
		},
		{
			name:     "backslash escape ignored",
			input:    "foo\\$(VAR_C)bar",
			expected: "foo\\$(VAR_C)bar",
		},
		{
			name:     "backslash escape ignored",
			input:    "foo\\\\$(VAR_C)bar",
			expected: "foo\\\\$(VAR_C)bar",
		},
		{
			name:     "lots of backslashes",
			input:    "foo\\\\\\\\$(VAR_A)bar",
			expected: "foo\\\\\\\\$(VAR_A)bar",
		},
		{
			name:     "nested var references",
			input:    "$(VAR_A$(VAR_B))",
			expected: "$(VAR_A$(VAR_B))",
		},
		{
			name:     "nested var references second type",
			input:    "$(VAR_A$(VAR_B)",
			expected: "$(VAR_A$(VAR_B)",
		},
		{
			name:     "value is a reference",
			input:    "$(VAR_REF)",
			expected: "$(VAR_A)",
			counts:   map[string]int{"VAR_REF": 1},
		},
		{
			name:     "value is a reference x 2",
			input:    "%%$(VAR_REF)--$(VAR_REF)%%",
			expected: "%%$(VAR_REF)--$(VAR_REF)%%",
		},
		{
			name:     "empty var",
			input:    "foo$(VAR_EMPTY)bar",
			expected: "foo$(VAR_EMPTY)bar",
		},
		{
			name:     "unterminated expression",
			input:    "foo$(VAR_Awhoops!",
			expected: "foo$(VAR_Awhoops!",
		},
		{
			name:     "expression without operator",
			input:    "f00__(VAR_A)__",
			expected: "f00__(VAR_A)__",
		},
		{
			name:     "shell special vars pass through",
			input:    "$?_boo_$!",
			expected: "$?_boo_$!",
		},
		{
			name:     "bare operators are ignored",
			input:    "$VAR_A",
			expected: "$VAR_A",
		},
		{
			name:     "undefined vars are passed through",
			input:    "$(VAR_DNE)",
			expected: "$(VAR_DNE)",
		},
		{
			name:     "multiple (even) operators, var undefined",
			input:    "$$$$$$(BIG_MONEY)",
			expected: "$$$$$$(BIG_MONEY)",
		},
		{
			name:     "multiple (even) operators, var defined",
			input:    "$$$$$$(VAR_A)",
			expected: "$$$$$$(VAR_A)",
		},
		{
			name:     "multiple (odd) operators, var undefined",
			input:    "$$$$$$$(GOOD_ODDS)",
			expected: "$$$$$$$(GOOD_ODDS)",
		},
		{
			name:     "multiple (odd) operators, var defined",
			input:    "$$$$$$$(VAR_A)",
			expected: "$$$$$$$(VAR_A)",
		},
		{
			name:     "missing open expression",
			input:    "$VAR_A)",
			expected: "$VAR_A)",
		},
		{
			name:     "shell syntax ignored",
			input:    "${VAR_A}",
			expected: "${VAR_A}",
		},
		{
			name:     "trailing incomplete expression not consumed",
			input:    "$(VAR_B)_______$(A",
			expected: "$(VAR_B)_______$(A",
		},
		{
			name:     "trailing incomplete expression, no content, is not consumed",
			input:    "$(VAR_C)_______$(",
			expected: "$(VAR_C)_______$(",
		},
		{
			name:     "operator at end of input string is preserved",
			input:    "$(VAR_A)foobarzab$",
			expected: "$(VAR_A)foobarzab$",
		},
		{
			name:     "shell escaped incomplete expr",
			input:    "foo-\\$(VAR_A",
			expected: "foo-\\$(VAR_A",
		},
		{
			name:     "lots of $( in middle",
			input:    "--$($($($($--",
			expected: "--$($($($($--",
		},
		{
			name:     "lots of $( in beginning",
			input:    "$($($($($--foo$(",
			expected: "$($($($($--foo$(",
		},
		{
			name:     "lots of $( at end",
			input:    "foo0--$($($($(",
			expected: "foo0--$($($($(",
		},
		{
			name:     "escaped operators in variable names are not escaped",
			input:    "$(foo$$var)",
			expected: "$(foo$$var)",
		},
		{
			name:     "newline not inlined",
			input:    "\n",
			expected: "\n",
		},
	}

	for _, tc := range cases {
		counts := make(map[string]int)
		mapping := InlineFuncFor(counts, context...)
		inlined := Inline(tc.input, mapping)
		if e, a := tc.expected, inlined; e != a {
			t.Errorf("%v: expected %q, got %q", tc.name, e, a)
		}
		if len(counts) != len(tc.counts) {
			t.Errorf("%v: len(counts)=%d != len(tc.counts)=%d",
				tc.name, len(counts), len(tc.counts))
		}
		if len(tc.counts) > 0 {
			for k, expectedCount := range tc.counts {
				c, ok := counts[k]
				if ok {
					if c != expectedCount {
						t.Errorf(
							"%v: k=%s, expected count %d, got %d",
							tc.name, k, expectedCount, c)
					}
				} else {
					t.Errorf(
						"%v: k=%s, expected count %d, got zero",
						tc.name, k, expectedCount)
				}
			}
		}
	}
}
