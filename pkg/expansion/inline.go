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

// Package expansion provides functions find and replace $(FOO) style variables in strings.
package expansion

// InlineFuncFor returns a inline function for use with Inline that
// implements the expansion semantics defined in the inline spec; it
// returns the input string wrapped in the expansion syntax if no mapping
// for the input is found.
func InlineFuncFor(
	counts map[string]int,
	context ...map[string]interface{}) func(string) interface{} {
	return func(input string) interface{} {
		for _, vars := range context {
			val, ok := vars[input]
			if ok {
				counts[input]++
				return val
			}
		}
		return syntaxWrap(input)
	}
}

// Expand replaces variable references in the input string according to
// the expansion spec using the given mapping function to resolve the
// values of variables.
func Inline(input string, inline func(string) interface{}) interface{} {

	if input[0] != operator {
		// This is not the right syntax for an inline
		return input
	}

	read, isVar, _ := tryReadVariableName(input[1:])

	if isVar && input == syntaxWrap(read) {
		// We were able to read a variable name correctly;
		// apply the mapping to the variable name and
		// return the object.
		return inline(read)
	}

	// This is not the right syntax for an inline
	return input
}
