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

// Package expansion provides functions for finding and replacing $(FOO) style
// variables in strings.
package expansion

import (
	"strings"
)

const (
	Dot string = "."
)

// Skeleton of the method to detect if a variable name
// is candidate for var and reference autoconfiguration
func matchAutoConfigPattern(detectedName string) bool {
	if strings.ContainsAny(detectedName, ")$({}") {
		return false
	}

	s := strings.Split(detectedName, Dot)
	if len(s) < 3 {
		return false
	}

	return true
}

// Detect detects variable references in the input string.
func Detect(input string) []string {
	detectedVars := []string{}
	for cursor := 0; cursor < len(input); cursor++ {
		if input[cursor] == operator && cursor+1 < len(input) {
			// Attempt to read the variable name as defined by the
			// syntax from the input string
			read, isVar, advance := tryReadVariableName(input[cursor+1:])

			if isVar && matchAutoConfigPattern(read) {
				// We were able to read a variable name correctly;
				detectedVars = append(detectedVars, read)
			}

			// Advance the cursor in the input string to account for
			// bytes consumed to read the variable name expression
			cursor += advance
		}
	}
	return detectedVars
}
