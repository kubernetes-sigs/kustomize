// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	var testCases = map[string]struct {
		raw    string
		v      SemVer
		errMsg string
	}{
		"one": {
			raw:    "v1.2.3",
			v:      SemVer{major: 1, minor: 2, patch: 3},
			errMsg: "",
		},
		"two": {
			raw:    "v2.0.9999",
			v:      SemVer{major: 2, minor: 0, patch: 9999},
			errMsg: "",
		},
		"three": {
			raw:    "pizza",
			v:      zero,
			errMsg: "\"pizza\" too short to be a version",
		},
		"non-digit": {
			raw:    "v1.x.222",
			v:      zero,
			errMsg: "strconv.Atoi: parsing \"x\": invalid syntax",
		},
		"bad fields": {
			raw:    "v1.222",
			v:      zero,
			errMsg: "\"v1.222\" doesn't have the form v1.2.3",
		},
	}
	for n, tc := range testCases {
		v, err := Parse(tc.raw)
		if err == nil {
			if tc.errMsg != "" {
				t.Errorf(
					"%s: no error, but expected err %q", n, tc.errMsg)
			}
			if !v.Equals(tc.v) {
				t.Errorf(
					"%s: expected %v, got %v", n, tc.v, v)
			}
		} else {
			if tc.errMsg == "" {
				t.Errorf(
					"%s: unexpected error %v", n, err)
			} else {
				if tc.errMsg != err.Error() {
					t.Errorf(
						"%s: expected err msg %q, but got %q",
						n, tc.errMsg, err.Error())
				}
			}
		}
	}
}

func TestLessThan(t *testing.T) {
	var testCases = map[string]struct {
		v1       SemVer
		v2       SemVer
		expected bool
	}{
		"one": {
			v1:       SemVer{major: 2, minor: 2, patch: 3},
			v2:       SemVer{major: 1, minor: 2, patch: 3},
			expected: false,
		},
		"two": {
			v1:       SemVer{major: 1, minor: 3, patch: 3},
			v2:       SemVer{major: 1, minor: 2, patch: 3},
			expected: false,
		},
		"three": {
			v1:       SemVer{major: 1, minor: 2, patch: 4},
			v2:       SemVer{major: 1, minor: 2, patch: 3},
			expected: false,
		},
		"eq": {
			v1:       SemVer{major: 2, minor: 2, patch: 3},
			v2:       SemVer{major: 2, minor: 2, patch: 3},
			expected: false,
		},
		"four": {
			v1:       zero,
			v2:       SemVer{major: 0, minor: 0, patch: 1},
			expected: true,
		},
	}
	for n, tc := range testCases {
		actual := tc.v1.LessThan(tc.v2)
		if actual != tc.expected {
			t.Errorf(
				"%s: expected %v, got %v for %s LessThan %s",
				n, tc.expected, actual, tc.v1.String(), tc.v2.String())
		}
	}
}
