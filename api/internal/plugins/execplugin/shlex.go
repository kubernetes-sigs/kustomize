// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin

import (
	"fmt"
	"strings"
	"unicode"
)

// ShlexSplit splits a string into a slice of strings using shell-style rules for quoting and commenting
// Similar to Python's shlex.split with comments enabled
func ShlexSplit(s string) ([]string, error) {
	return shlexSplit(s)
}

func shlexSplit(s string) ([]string, error) {
	result := []string{}

	var current strings.Builder
	var quote rune
	var escaped bool

	for _, r := range s {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\' && quote != '\'':
			escaped = true
		case (r == '\'' || r == '"') && quote == 0:
			quote = r
		case r == quote:
			quote = 0
		case r == '#' && quote == 0:
			// Comment starts, ignore the rest of the line
			if current.Len() > 0 {
				result = append(result, current.String())
			}
			return result, nil
		case quote == 0 && unicode.IsSpace(r):
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote in string")
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result, nil
}
