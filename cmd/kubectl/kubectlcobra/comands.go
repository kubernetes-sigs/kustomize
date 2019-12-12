// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"flag"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/diff"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/kustomize/kyaml/commandutil"

	// initialize auth
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// GetCommand returns a command from kubectl to install
func GetCommand(parent *cobra.Command) *cobra.Command {
	if !commandutil.GetAlphaEnabled() {
		return &cobra.Command{
			Use:   "resources",
			Short: "[Alpha] To enable set KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true",
			Long:  "[Alpha] To enable set KUSTOMIZE_ENABLE_ALPHA_COMMANDS=true",
		}
	}

	r := &cobra.Command{
		Use:   "resources",
		Short: "[Alpha] Perform cluster operations using declarative configuration",
		Long:  "[Alpha] Perform cluster operations using declarative configuration",
	}

	// configure kubectl dependencies and flags
	flags := r.Flags()
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := util.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(r.PersistentFlags())
	r.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	f := util.NewFactory(matchVersionKubeConfigFlags)

	var ioStreams genericclioptions.IOStreams

	if parent != nil {
		ioStreams.In = parent.InOrStdin()
		ioStreams.Out = parent.OutOrStdout()
		ioStreams.ErrOut = parent.ErrOrStderr()
	} else {
		ioStreams.In = os.Stdin
		ioStreams.Out = os.Stdout
		ioStreams.ErrOut = os.Stderr
	}

	names := []string{"apply", "diff"}
	applyCmd := apply.NewCmdApply("kustomize", f, ioStreams)
	updateHelp(names, applyCmd)
	diffCmd := diff.NewCmdDiff(f, ioStreams)
	updateHelp(names, diffCmd)

	r.AddCommand(applyCmd, diffCmd)
	return r
}

// updateHelp replaces `kubectl` help messaging with `kustomize` help messaging
func updateHelp(names []string, c *cobra.Command) {
	for i := range names {
		name := names[i]
		c.Short = strings.ReplaceAll(c.Short, "kubectl "+name, "kustomize "+name)
		c.Long = strings.ReplaceAll(c.Long, "kubectl "+name, "kustomize "+name)
		c.Example = strings.ReplaceAll(c.Example, "kubectl "+name, "kustomize "+name)
	}
}
