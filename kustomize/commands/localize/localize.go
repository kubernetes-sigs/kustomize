// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localize

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/localizer"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var theArgs struct {
	target string
	newDir string
}

var theFlags struct {
	scope string
}

func NewCmdLocalize(fSys filesys.FileSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "localize",
		Short: "creates localized copy of target kustomization root",
		Long: "creates copy of target kustomization root and overwrites remote references with local paths to downloads of " +
			"referenced files",
		Example: `
	# create localized copy of target at newDir and target cannot references files outside of scope
	kustomize localize <target> <newDir> --scope <scope>

	# create localized copy of target at newDir and target cannot references files outside of itself
	kustomize localize <target> <newDir>
`,
		RunE: func(command *cobra.Command, args []string) error {
			if err := validate(args); err != nil {
				return err
			}
			return localizer.Run(fSys, theArgs.target, theFlags.scope, theArgs.newDir)
		},
	}
	AddFlagScope(cmd.Flags())
	return cmd
}

func validate(args []string) error {
	// TODO: think about argument edge cases
	switch len(args) {
	case 0:
		return errors.Errorf("need at least 1 target argument to localize")
	case 1:
		return errors.Errorf("I have not yet written support for the default newDir")
	case 2:
		theArgs.target = args[0]
		theArgs.newDir = args[1]
	default:
		return errors.WrapPrefixf("too many arguments", "%b arguments", len(args))
	}
	return validateFlagScope()
}
