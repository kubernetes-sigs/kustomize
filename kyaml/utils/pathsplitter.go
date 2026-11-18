// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils

import "strings"

// TODO: Move these to kyaml

// PathSplitter splits a delimited string, permitting escaped delimiters.
func PathSplitter(path string, delimiter string) []string {
	ps := strings.Split(path, delimiter)
	var res []string

	// allow path to start with forward slash
	// i.e. /a/b/c
	if len(ps) > 1 && ps[0] == "" {
		ps = ps[1:]
	}

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

// SmarterPathSplitter splits a path, retaining bracketed elements.
// If the element is a list entry identifier (defined by the '='),
// it will retain the brackets.
// E.g. "[name=com.foo.someapp]" survives as one thing after splitting
// "spec.template.spec.containers.[name=com.foo.someapp].image"
// See kyaml/yaml/match.go for use of list entry identifiers.
// If the element is a mapping entry identifier, it will remove the
// brackets.
// E.g. "a.b.c" survives as one thing after splitting
// "metadata.annotations.[a.b.c]
// This function uses `PathSplitter`, so it also respects escaped delimiters.
func SmarterPathSplitter(path string, delimiter string) []string {
	var result []string
	split := PathSplitter(path, delimiter)

	for i := 0; i < len(split); i++ {
		elem := split[i]
		switch {
		case strings.HasPrefix(elem, "[") && !strings.HasSuffix(elem, "]"):
			// The bracketed element contains the delimiter, so it was split
			// across multiple parts. Continue until we find the matching "]".
			bracketed := []string{elem}
			for i < len(split)-1 {
				i++
				bracketed = append(bracketed, split[i])
				if strings.HasSuffix(split[i], "]") {
					break
				}
			}
			result = append(result, classifyBracketed(strings.Join(bracketed, delimiter)))
		case len(elem) > 1 && strings.HasPrefix(elem, "[") && strings.HasSuffix(elem, "]"):
			// The bracketed element is self-contained (no delimiter inside),
			// e.g. "[name=first]" or "[mykey]".
			result = append(result, classifyBracketed(elem))
		default:
			result = append(result, elem)
		}
	}
	return result
}

// classifyBracketed decides how to treat a fully bracketed path element.
// If it is a list entry identifier (contains '='), the brackets are retained.
// Otherwise it is a mapping entry identifier and the brackets are removed.
func classifyBracketed(bracketed string) string {
	if strings.Contains(bracketed, "=") {
		return bracketed
	}
	return strings.Trim(bracketed, "[]")
}
