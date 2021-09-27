// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
package merge2

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Test_determineSmpDirective(t *testing.T) {
	var cases = map[string]struct {
		patch       string
		elided      string
		expected    smpDirective
		errExpected string
	}{
		`scalar`: {
			patch:       "dumb",
			expected:    smpMerge,
			errExpected: "no implemented strategic merge patch strategy",
		},
		`list merge`: {
			patch: `
- one
- two
- three
- $patch: merge
`,
			expected: smpMerge,
			elided: `- one
- two
- three
`,
		},
		"list replace": {
			patch: `
- one
- two
- three
- $patch: replace
`,
			expected: smpReplace,
			elided: `- one
- two
- three
`,
		},
		"list delete": {
			patch: `
- one
- two
- three
- $patch: delete
`,
			expected: smpDelete,
			elided: `- one
- two
- three
`,
		},
		"list default": {
			patch: `
- one
- two
- three
`,
			expected: smpMerge,
			elided: `- one
- two
- three
`,
		},
		`map replace`: {
			patch: `
metal: heavy
$patch: replace
veggie: carrot
`,
			expected: smpReplace,
			elided: `metal: heavy
veggie: carrot
`,
		},
		`map delete`: {
			patch: `
metal: heavy
$patch: delete
veggie: carrot
`,
			expected: smpDelete,
			elided: `metal: heavy
veggie: carrot
`,
		},
		`map merge`: {
			patch: `
metal: heavy
$patch: merge
veggie: carrot
`,
			expected: smpMerge,
			elided: `metal: heavy
veggie: carrot
`,
		},
		`map default`: {
			patch: `
metal: heavy
veggie: carrot
`,
			expected: smpMerge,
			elided: `metal: heavy
veggie: carrot
`,
		},
	}

	for n := range cases {
		tc := cases[n]
		t.Run(n, func(t *testing.T) {
			p, err := yaml.Parse(tc.patch)
			if err != nil {
				t.Fatalf("unexpected parse err %v", err)
			}
			unwrapped := yaml.NewRNode(p.YNode())
			actual, err := determineSmpDirective(unwrapped)
			if err == nil {
				if tc.errExpected != "" {
					t.Fatalf("should have seen an error")
				}
				if tc.expected != actual {
					t.Fatalf("expected %s, got %s", tc.expected, actual)
				}
				if tc.elided != unwrapped.MustString() {
					t.Fatalf(
						"expected %s, got %s",
						tc.elided, unwrapped.MustString())
				}
			} else {
				if tc.errExpected == "" {
					t.Fatalf("unexpected err: %v", err)
				}
				if !strings.Contains(err.Error(), tc.errExpected) {
					t.Fatalf("expected some error other than:  %v", err)
				}
			}
		})
	}
}
