// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/pathutil"
)

// cmdRunner interface holds executeCmd definition which executes respective command's
// implementation on single package
type cmdRunner interface {
	executeCmd(w io.Writer, pkgPath string) error
}

// executeCmdOnPkgs struct holds the parameters necessary to
// execute the filter command on packages in rootPkgPath
type executeCmdOnPkgs struct {
	rootPkgPath        string
	recurseSubPackages bool
	needOpenAPI        bool
	cmdRunner          cmdRunner
	writer             io.Writer
	skipPkgPathPrint   bool
}

// executeCmdOnPkgs takes the function definition for a command to be executed on single package, applies that definition
// recursively on all the subpackages present in rootPkgPath if recurseSubPackages is true, else applies the command on rootPkgPath only
func (e executeCmdOnPkgs) execute() error {
	pkgsPaths, err := pathutil.DirsWithFile(e.rootPkgPath, ext.KRMFileName(), e.recurseSubPackages)
	if err != nil {
		return err
	}

	if len(pkgsPaths) == 0 {
		// at this point, there are no openAPI files in the rootPkgPath
		if e.needOpenAPI {
			// few executions need openAPI file to be present(ex: setters commands), if true throw an error
			return errors.Errorf("unable to find %q in package %q", ext.KRMFileName(), e.rootPkgPath)
		}

		// add root path for commands which doesn't need openAPI(ex: annotate, fmt)
		pkgsPaths = []string{e.rootPkgPath}
	}

	for i := range pkgsPaths {
		pkgPath := pkgsPaths[i]
		// Add schema present in openAPI file for current package
		if e.needOpenAPI {
			if err := openapi.AddSchemaFromFile(filepath.Join(pkgPath, ext.KRMFileName())); err != nil {
				return err
			}
		}

		if !e.skipPkgPathPrint {
			fmt.Fprintf(e.writer, "%s/\n", pkgPath)
		}

		err := e.cmdRunner.executeCmd(e.writer, pkgPath)
		if err != nil {
			return err
		}

		if i != len(pkgsPaths)-1 {
			fmt.Fprint(e.writer, "\n")
		}

		// Delete schema present in openAPI file for current package
		if e.needOpenAPI {
			if err := openapi.DeleteSchemaInFile(filepath.Join(pkgPath, ext.KRMFileName())); err != nil {
				return err
			}
		}
	}
	return nil
}

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
			fmt.Fprintf(os.Stderr, "%s", err.Stack())
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
