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
	"reflect"
	"testing"

	. "sigs.k8s.io/kustomize/v3/pkg/expansion"
)

var Empty = []string{}

func TestDetection(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "missing name",
			input:    "$(KindA)",
			expected: Empty,
		},
		{
			name:     "missing path",
			input:    "$(KindA.name-a)",
			expected: Empty,
		},
		{
			name:     "whole string",
			input:    "$(KindA.name-a.path.a)",
			expected: []string{"KindA.name-a.path.a"},
		},
		{
			name:     "repeat",
			input:    "$(KindA.name-a.path.a)-$(KindA.name-a.path.a)",
			expected: []string{"KindA.name-a.path.a", "KindA.name-a.path.a"},
		},
		{
			name:     "multiple repeats",
			input:    "$(KindA.name-a.path.a)-$(KindB.name-b.path.b)-$(KindB.name-b.path.b)-$(KindB.name-b.path.b)-$(KindA.name-a.path.a)",
			expected: []string{"KindA.name-a.path.a", "KindB.name-b.path.b", "KindB.name-b.path.b", "KindB.name-b.path.b", "KindA.name-a.path.a"},
		},
		{
			name:     "beginning",
			input:    "$(KindA.name-a.path.a)-1",
			expected: []string{"KindA.name-a.path.a"},
		},
		{
			name:     "middle",
			input:    "___$(KindB.name-b.path.b)___",
			expected: []string{"KindB.name-b.path.b"},
		},
		{
			name:     "end",
			input:    "___$(KindC.name-c.path.c)",
			expected: []string{"KindC.name-c.path.c"},
		},
		{
			name:     "compound",
			input:    "$(KindA.name-a.path.a)_$(KindB.name-b.path.b)_$(KindC.name-c.path.c)",
			expected: []string{"KindA.name-a.path.a", "KindB.name-b.path.b", "KindC.name-c.path.c"},
		},
		{
			name:     "escape & expand",
			input:    "$$(KindB.name-b.path.b)_$(KindA.name-a.path.a)",
			expected: []string{"KindA.name-a.path.a"},
		},
		{
			name:     "compound escape",
			input:    "$$(KindA.name-a.path.a)_$$(KindB.name-b.path.b)",
			expected: Empty,
		},
		{
			name:     "mixed in escapes",
			input:    "f000-$$KindA.name-a.path.a",
			expected: Empty,
		},
		{
			name:     "backslash escape ignored",
			input:    "foo\\$(KindC.name-c.path.c)bar",
			expected: []string{"KindC.name-c.path.c"},
		},
		{
			name:     "backslash escape ignored",
			input:    "foo\\\\$(KindC.name-c.path.c)bar",
			expected: []string{"KindC.name-c.path.c"},
		},
		{
			name:     "lots of backslashes",
			input:    "foo\\\\\\\\$(KindA.name-a.path.a)bar",
			expected: []string{"KindA.name-a.path.a"},
		},
		{
			name:     "nested var references",
			input:    "$(KindA.name-a.path.a$(KindB.name-b.path.b))",
			expected: Empty,
		},
		{
			name:     "nested var references second type",
			input:    "$(KindA.name-a.path.a$(KindB.name-b.path.b)",
			expected: Empty,
		},
		{
			name:     "value is a reference",
			input:    "$(VAR_REF)",
			expected: Empty,
		},
		{
			name:     "value is a reference x 2",
			input:    "%%$(VAR_REF)--$(VAR_REF)%%",
			expected: Empty,
		},
		{
			name:     "empty var",
			input:    "foo$(VAR_EMPTY)bar",
			expected: Empty,
		},
		{
			name:     "unterminated expression",
			input:    "foo$(KindA.name-a.path.awhoops!",
			expected: Empty,
		},
		{
			name:     "expression without operator",
			input:    "f00__(KindA.name-a.path.a)__",
			expected: Empty,
		},
		{
			name:     "shell special vars pass through",
			input:    "$?_boo_$!",
			expected: Empty,
		},
		{
			name:     "bare operators are ignored",
			input:    "$KindA.name-a.path.a",
			expected: Empty,
		},
		{
			name:     "undefined vars are passed through",
			input:    "$(VAR_DNE)",
			expected: Empty,
		},
		{
			name:     "multiple (even) operators, var undefined",
			input:    "$$$$$$(BIG_MONEY)",
			expected: Empty,
		},
		{
			name:     "multiple (even) operators, var defined",
			input:    "$$$$$$(KindA.name-a.path.a)",
			expected: Empty,
		},
		{
			name:     "multiple (odd) operators, var undefined",
			input:    "$$$$$$$(GOOD_ODDS)",
			expected: Empty,
		},
		{
			name:     "multiple (odd) operators, var defined",
			input:    "$$$$$$$(KindA.name-a.path.a)",
			expected: []string{"KindA.name-a.path.a"},
		},
		{
			name:     "missing open expression",
			input:    "$KindA.name-a.path.a)",
			expected: Empty,
		},
		{
			name:     "shell syntax ignored",
			input:    "${KindA.name-a.path.a}",
			expected: Empty,
		},
		{
			name:     "trailing incomplete expression not consumed",
			input:    "$(KindB.name-b.path.b)_______$(A",
			expected: []string{"KindB.name-b.path.b"},
		},
		{
			name:     "trailing incomplete expression, no content, is not consumed",
			input:    "$(KindC.name-c.path.c)_______$(",
			expected: []string{"KindC.name-c.path.c"},
		},
		{
			name:     "operator at end of input string is preserved",
			input:    "$(KindA.name-a.path.a)expectedzab$",
			expected: []string{"KindA.name-a.path.a"},
		},
		{
			name:     "shell escaped incomplete expr",
			input:    "foo-\\$(KindA.name-a.path.a",
			expected: Empty,
		},
		{
			name:     "lots of $( in middle",
			input:    "--$($($($($--",
			expected: Empty,
		},
		{
			name:     "lots of $( in beginning",
			input:    "$($($($($--foo$(",
			expected: Empty,
		},
		{
			name:     "lots of $( at end",
			input:    "foo0--$($($($(",
			expected: Empty,
		},
		{
			name:     "escaped operators in variable names are not escaped",
			input:    "$(foo$$var)",
			expected: Empty,
		},
		{
			name:     "newline not expanded",
			input:    "\n",
			expected: Empty,
		},
	}

	for _, tc := range cases {
		detected := Detect(fmt.Sprintf("%v", tc.input))
		if !reflect.DeepEqual(tc.expected, detected) {
			t.Errorf("%v: expected %q, got %q", tc.name, tc.expected, detected)
		}
	}
}
