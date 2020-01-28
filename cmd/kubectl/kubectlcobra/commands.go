// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/diff"
	"k8s.io/kubectl/pkg/cmd/util"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"sigs.k8s.io/kustomize/kyaml/commandutil"
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
	applyCmd := NewCmdApply("kustomize", f, ioStreams)
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

// NewCmdApply creates the `apply` command
func NewCmdApply(baseName string, f util.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	applier := newApplier(f, ioStreams)
	printer := &BasicPrinter{
		ioStreams: ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "apply (-f FILENAME | -k DIRECTORY)",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Apply a configuration to a resource by filename or stdin"),
		//Long:                  applyLong,
		//Example:               applyExample,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				// check is kustomize, if so update
				applier.applyOptions.DeleteFlags.FileNameFlags.Kustomize = &args[0]
			}

			cmdutil.CheckErr(applier.Initialize(cmd))

			// Create a context with the provided timout from the cobra parameter.
			ctx, cancel := context.WithTimeout(context.Background(), applier.statusOptions.timeout)
			defer cancel()
			// Run the applier. It will return a channel where we can receive updates
			// to keep track of progress and any issues.
			ch := applier.Run(ctx)

			// The printer will print updates from the channel. It will block
			// until the channel is closed.
			printer.Print(ch)
		},
	}

	applier.SetFlags(cmd)

	cmdutil.AddValidateFlags(cmd)
	cmd.Flags().BoolVar(&applier.applyOptions.ServerDryRun, "server-dry-run", applier.applyOptions.ServerDryRun, "If true, request will be sent to server with dry-run flag, which means the modifications won't be persisted. This is an alpha feature and flag.")
	cmd.Flags().Bool("dry-run", false, "If true, only print the object that would be sent, without sending it. Warning: --dry-run cannot accurately output the result of merging the local manifest and the server-side data. Use --server-dry-run to get the merged result instead.")
	cmdutil.AddServerSideApplyFlags(cmd)

	return cmd
}

// PrependGroupingObject orders the objects to apply so the "grouping"
// object stores the inventory, and it is first to be applied.
func PrependGroupingObject(o *apply.ApplyOptions) func() error {
	return func() error {
		if o == nil {
			return fmt.Errorf("ApplyOptions are nil")
		}
		infos, err := o.GetObjects()
		if err != nil {
			return err
		}
		_, exists := findGroupingObject(infos)
		if exists {
			if err := addInventoryToGroupingObj(infos); err != nil {
				return err
			}
			if !sortGroupingObject(infos) {
				return err
			}
		}
		return nil
	}
}

// Prune deletes previously applied objects that have been
// omitted in the current apply. The previously applied objects
// are reached through ConfigMap grouping objects.
func prune(f util.Factory, o *apply.ApplyOptions) func() error {
	return func() error {
		po, err := NewPruneOptions(f, o)
		if err != nil {
			return err
		}
		return po.Prune()
	}
}
