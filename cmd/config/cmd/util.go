// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
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

func handleError(c *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	if StackOnError {
		if err, ok := err.(*errors.Error); ok {
			fmt.Fprint(os.Stderr, fmt.Sprintf("%s", err.Stack()))
		}
	}

	if ExitOnError {
		fmt.Fprintf(c.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
	}
	return err
}

// ExitOnError if true, will cause commands to call os.Exit instead of returning an error.
// Used for skipping printing usage on failure.
var ExitOnError bool

// StackOnError if true, will print a stack trace on failure.
var StackOnError bool

const cmdName = "kustomize config"

// FixDocs replaces instances of old with new in the docs for c
func fixDocs(new string, c *cobra.Command) {
	c.Use = strings.ReplaceAll(c.Use, cmdName, new)
	c.Short = strings.ReplaceAll(c.Short, cmdName, new)
	c.Long = strings.ReplaceAll(c.Long, cmdName, new)
	c.Example = strings.ReplaceAll(c.Example, cmdName, new)
}
