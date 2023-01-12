// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localize

import (
	"fmt"
	"io"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	lclzr "sigs.k8s.io/kustomize/api/krusty/localizer"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const cmdName = "localize"
const numArgs = 2
const alphaWarning = `Warning: This is currently an alpha command.
`

type arguments struct {
	target string
	dest   string
}

type flags struct {
	scope string
}

// NewCmdLocalize returns a new localize command.
func NewCmdLocalize(fs filesys.FileSystem, writer io.Writer) *cobra.Command {
	log.SetOutput(writer)
	var f flags
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [target [destination]]", cmdName),
		Short: "Creates localized copy of target kustomization root at destination",
		Long: fmt.Sprintf(`Creates copy of target kustomization directory
or versioned URL at destination, where remote references in the original are 
replaced by local references to the downloaded remote content.

If target is not specified, the current working directory will be used. 
Destination is a path to a new directory in an existing directory. If 
destination is not specified, a new directory will be created in the current 
working directory. 

For details, see: https://kubectl.docs.kubernetes.io/references/kustomize/cmd/

Disclaimer:
This command does not yet localize helm or KRM plugin fields. This command also
alphabetizes kustomization fields in the localized copy.

%s`, alphaWarning),
		Example: fmt.Sprintf(`
# Localize the current working directory, with default scope and destination
%s %s 

# Localize some local directory, with scope and default destination
%s %s /home/path/scope/target --scope /home/path/scope

# Localize remote at set destination relative to working directory
%s %s https://github.com/kubernetes-sigs/kustomize//api/krusty/testdata/localize/simple?ref=v4.5.7 path/non-existing-dir
`, konfig.ProgramName, cmdName, konfig.ProgramName, cmdName, konfig.ProgramName, cmdName),
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(numArgs),
		RunE: func(cmd *cobra.Command, rawArgs []string) error {
			_, _ = writer.Write([]byte(alphaWarning))
			args := matchArgs(rawArgs)
			return errors.Wrap(lclzr.Run(fs, args.target, f.scope, args.dest))
		},
	}
	AddFlagScope(&f, cmd.Flags())
	return cmd
}

// matchArgs matches user-entered userArgs, which cannot exceed max length, with
// arguments.
func matchArgs(rawArgs []string) arguments {
	var args arguments
	switch len(rawArgs) {
	case numArgs:
		args.dest = rawArgs[1]
		fallthrough
	case 1:
		args.target = rawArgs[0]
	case 0:
		args.target = filesys.SelfDir
	}
	return args
}
