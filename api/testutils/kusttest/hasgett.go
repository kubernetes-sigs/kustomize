// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/resmap"
)

type hasGetT interface {
	GetT() *testing.T
}

func assertActualEqualsExpectedWithTweak(
	ht hasGetT,
	m resmap.ResMap,
	tweaker func([]byte) []byte, expected string) {
	if m == nil {
		ht.GetT().Fatalf("Map should not be nil.")
	}
	// Ignore leading linefeed in expected value
	// to ease readability of tests.
	if len(expected) > 0 && expected[0] == 10 {
		expected = expected[1:]
	}
	actual, err := m.AsYaml()
	if err != nil {
		ht.GetT().Fatalf("Unexpected err: %v", err)
	}
	if tweaker != nil {
		actual = tweaker(actual)
	}
	if string(actual) != expected {
		reportDiffAndFail(ht.GetT(), actual, expected)
	}
}

// Pretty printing of file differences.
func reportDiffAndFail(
	t *testing.T, actual []byte, expected string) {
	sE, maxLen := convertToArray(expected)
	sA, _ := convertToArray(string(actual))
	fmt.Println("===== ACTUAL BEGIN ========================================")
	fmt.Print(string(actual))
	fmt.Println("===== ACTUAL END ==========================================")
	format := fmt.Sprintf("%%s  %%-%ds %%s\n", maxLen+4)
	limit := 0
	if len(sE) < len(sA) {
		limit = len(sE)
	} else {
		limit = len(sA)
	}
	fmt.Printf(format, " ", "EXPECTED", "ACTUAL")
	fmt.Printf(format, " ", "--------", "------")
	for i := 0; i < limit; i++ {
		fmt.Printf(format, hint(sE[i], sA[i]), sE[i], sA[i])
	}
	if len(sE) < len(sA) {
		for i := len(sE); i < len(sA); i++ {
			fmt.Printf(format, "X", "", sA[i])
		}
	} else {
		for i := len(sA); i < len(sE); i++ {
			fmt.Printf(format, "X", sE[i], "")
		}
	}
	t.Fatalf("Expected not equal to actual")
}

func hint(a, b string) string {
	if a == b {
		return " "
	}
	return "X"
}

func convertToArray(x string) ([]string, int) {
	a := strings.Split(strings.TrimSuffix(x, "\n"), "\n")
	maxLen := 0
	for i, v := range a {
		z := tabToSpace(v)
		if len(z) > maxLen {
			maxLen = len(z)
		}
		a[i] = z
	}
	return a, maxLen
}

func tabToSpace(input string) string {
	var result []string
	for _, i := range input {
		if i == 9 {
			result = append(result, "  ")
		} else {
			result = append(result, string(i))
		}
	}
	return strings.Join(result, "")
}
