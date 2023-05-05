// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localize

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	lclzr "sigs.k8s.io/kustomize/api/krusty/localizer"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const numArgs = 2

type arguments struct {
	target string
	dest   string
}

type theFlags struct {
	scope    string
	creds    string
	provider lclzr.SourceOCIProvider
}

// NewCmdLocalize returns a new localize command.
func NewCmdLocalize(fs filesys.FileSystem) *cobra.Command {
	var f theFlags
	f.provider.Set("generic")
	cmd := &cobra.Command{
		Use:   "localize [target [destination]]",
		Short: "[Alpha] Creates localized copy of target kustomization root at destination",
		Long: `[Alpha] Creates copy of target kustomization directory or 
versioned URL at destination, where remote references in the original 
are replaced by local references to the downloaded remote content.

If target is not specified, the current working directory will be used. 
Destination is a path to a new directory in an existing directory. If 
destination is not specified, a new directory will be created in the current 
working directory. 

For details, see: https://kubectl.docs.kubernetes.io/references/kustomize/cmd/

Disclaimer:
This command does not yet localize helm or KRM plugin fields. This command also
alphabetizes kustomization fields in the localized copy.
`,
		Example: `
# Localize the current working directory, with default scope and destination
kustomize localize 

# Localize some local directory, with scope and default destination
kustomize localize /home/path/scope/target --scope /home/path/scope

# Localize remote at set destination relative to working directory
kustomize localize https://github.com/kubernetes-sigs/kustomize//api/krusty/testdata/localize/simple?ref=v4.5.7 path/non-existing-dir

# Localize remote OCI manifest (if no folder is provided, the current folder is used)
kustomize localize oci://ghcr.io/my-user/oci-manifest:latest oci-manifest
`,
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(numArgs),
		RunE: func(cmd *cobra.Command, rawArgs []string) error {
			args := matchArgs(rawArgs)
			var dst string
			var err error
			// if it's an artifact download it
			if strings.HasPrefix(args.target, "oci://") {
				dst, err = lclzr.Pull(args.target, args.dest, f.provider, f.creds)
			} else {
				dst, err = lclzr.Run(fs, args.target, f.scope, args.dest)
			}
			if err != nil {
				return errors.Wrap(err)
			}

			log.Printf("SUCCESS: localized %q to directory %s\n", args.target, dst)
			return nil
		},
	}
	// no shorthand to avoid conflation with other flags
	cmd.Flags().StringVar(&f.scope,
		"scope",
		"",
		`Path to directory inside of which localize is limited to running.
Cannot specify for remote targets, as scope is by default the containing repo.
If not specified for local target, scope defaults to target.
`)
	cmd.Flags().StringVar(&f.creds, "creds", "", "credentials for OCI registry in the format <username>[:<password>] if --provider is generic")
	cmd.Flags().Var(&f.provider, "provider", f.provider.Description())
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
