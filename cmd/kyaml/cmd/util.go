// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package cmd

import (
	"fmt"
	"strings"
)

// parseFieldPath parse a flag value into a field path
func parseFieldPath(path string) ([]string, error) {
	// fixup '\.' so we don't split on it
	match := strings.ReplaceAll(path, "\\.", "$$$$")
	parts := strings.Split(match, ".")
	for i := range parts {
		parts[i] = strings.ReplaceAll(parts[i], "$$$$", ".")
	}

	// split the list index from the list field
	var newParts []string
	for i := range parts {
		if !strings.Contains(parts[i], "[") {
			newParts = append(newParts, parts[i])
			continue
		}
		p := strings.Split(parts[i], "[")
		if len(p) != 2 {
			return nil, fmt.Errorf("unrecognized path element: %s.  "+
				"Should be of the form 'list[field=value]'", parts[i])
		}
		p[1] = "[" + p[1]
		newParts = append(newParts, p[0], p[1])
	}
	return newParts, nil
}
