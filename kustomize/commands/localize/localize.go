// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localize

import (
	"bytes"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	lclzr "sigs.k8s.io/kustomize/api/krusty/localizer"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/build"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const numArgs = 2

type arguments struct {
	target string
	dest   string
}

type flags struct {
	scope    string
	noVerify bool
}

// NewCmdLocalize returns a new localize command.
func NewCmdLocalize(fs filesys.FileSystem) *cobra.Command {
	var f flags
	var buildBuffer bytes.Buffer
	buildCmd := build.NewCmdBuild(fs, &build.Help{}, &buildBuffer)
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
`,
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(numArgs),
		RunE: func(cmd *cobra.Command, rawArgs []string) error {
			args := matchArgs(rawArgs)
			dst, err := lclzr.Run(fs, args.target, f.scope, args.dest)
			if err != nil {
				return errors.Wrap(err)
			}

			if !f.noVerify {
				originalBuild, err := runBuildCmd(buildBuffer, buildCmd, args.target)
				if err != nil {
					return errors.Wrap(err)
				}

				buildDst := dst
				if f.scope != "" && f.scope != args.target {
					buildDst = filepath.Join(dst, filepath.Base(args.target))
				}

				localizedBuild, err := runBuildCmd(buildBuffer, buildCmd, buildDst)
				if err != nil {
					return errors.Wrap(err)
				}

				if localizedBuild != originalBuild {
					copyutil.PrettyFileDiff(originalBuild, localizedBuild)
					log.Fatalf("VERIFICATION FAILED: `kustomize build` for %s and %s are different after localization.\n", args.target, dst)
				}
				log.Printf("VERIFICATION SUCCESS: `kustomize build` for %s and %s are the same after localization.\n", args.target, dst)
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
	cmd.Flags().BoolVar(&f.noVerify,
		"no-verify",
		false,
		`Does not verify that the outputs of kustomize build for target and newDir are the same after localization.
		If not specified, this flag defaults to false and will run kustomize build.
	`)
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

func runBuildCmd(buffer bytes.Buffer, cmd *cobra.Command, folder string) (buildOutput string, err error) {
	buffer.Reset()
	buildErr := cmd.RunE(cmd, []string{folder})
	if buildErr != nil {
		log.Printf("If your target directory requires flags to build: \n"+
			"1. Add executable permissions for the downloaded exec binaries in '%s'. \n"+
			"2. Run kustomize build with the necessary flags and self-verify the outputs.", folder)
		return "", errors.Wrap(buildErr)
	}

	return buffer.String(), nil
}
