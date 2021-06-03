// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils

import "strings"

// TODO: Move these to kyaml

// PathSplitter splits a delimited string, permitting escaped delimiters.
func PathSplitter(path string, delimiter string) []string {
	ps := strings.Split(path, delimiter)
	var res []string
	res = append(res, ps[0])
	for i := 1; i < len(ps); i++ {
		last := len(res) - 1
		if strings.HasSuffix(res[last], `\`) {
			res[last] = strings.TrimSuffix(res[last], `\`) + delimiter + ps[i]
		} else {
			res = append(res, ps[i])
		}
	}
	return res
}

// SmarterPathSplitter splits a path, retaining bracketed list entry identifiers.
// E.g. [name=com.foo.someapp] survives as one thing after splitting
// "spec.template.spec.containers.[name=com.foo.someapp].image"
// See kyaml/yaml/match.go for use of list entry identifiers.
// This function uses `PathSplitter`, so it respects list entry identifiers
// and escaped delimiters.
func SmarterPathSplitter(path string, delimiter string) []string {
	var result []string
	split := PathSplitter(path, delimiter)

	for i := 0; i < len(split); i++ {
		elem := split[i]
		if strings.HasPrefix(elem, "[") && !strings.HasSuffix(elem, "]") {
			// continue until we find the matching "]"
			bracketed := []string{elem}
			for i < len(split)-1 {
				i++
				bracketed = append(bracketed, split[i])
				if strings.HasSuffix(split[i], "]") {
					break
				}
			}
			result = append(result, strings.Join(bracketed, delimiter))
		} else {
			result = append(result, elem)
		}
	}
	return result
}
