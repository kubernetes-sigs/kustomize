// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	o := apply.NewApplyOptions(ioStreams)
	so := newStatusOptions(f, ioStreams)
	o.PreProcessorFn = PrependGroupingObject(o)

	cmd := &cobra.Command{
		Use:                   "apply (FILENAME | DIRECTORY)",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Apply a configuration to a resource by filename or stdin"),
		//Long:                  applyLong,
		//Example:               applyExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(setFileNameFlags(args, o))

			cmdutil.CheckErr(o.Complete(f, cmd))
			cmdutil.CheckErr(o.Run())
			infos, _ := o.GetObjects()
			if so.wait {
				cmdutil.CheckErr(so.waitForStatus(infos))
			}
		},
	}

	// bind flag structs
	o.DeleteFlags.AddFlags(cmd)
	o.RecordFlags.AddFlags(cmd)
	o.PrintFlags.AddFlags(cmd)
	so.AddFlags(cmd)

	o.Overwrite = true

	cmdutil.AddValidateFlags(cmd)
	cmd.Flags().BoolVar(&o.ServerDryRun, "server-dry-run", o.ServerDryRun, "If true, request will be sent to server with dry-run flag, which means the modifications won't be persisted. This is an alpha feature and flag.")
	cmd.Flags().Bool("dry-run", false, "If true, only print the object that would be sent, without sending it. Warning: --dry-run cannot accurately output the result of merging the local manifest and the server-side data. Use --server-dry-run to get the merged result instead.")
	cmdutil.AddServerSideApplyFlags(cmd)

	// Hide flags that are set inside ApplyOptions, but that we don't really want exposed.
	cmdutil.CheckErr(cmd.Flags().MarkHidden("kustomize"))
	cmdutil.CheckErr(cmd.Flags().MarkHidden("filename"))
	cmdutil.CheckErr(cmd.Flags().MarkHidden("recursive"))

	return cmd
}

func setFileNameFlags(args []string, o *apply.ApplyOptions) error {
	// No arguments means we are reading from StdIn
	if len(args) == 0 {
		fileNames := []string{"-"}
		o.DeleteFlags.FileNameFlags.Filenames = &fileNames
		return nil
	}

	// If there are more than one argument, it can't be a kustomization,
	// so just set the FileNames argument. Also always set the Recursive flag
	// to true
	t := true
	if len(args) > 1 {
		o.DeleteFlags.FileNameFlags.Filenames = &args
		o.DeleteFlags.FileNameFlags.Recursive = &t
		return nil
	}

	// If we have exactly one parameters, we check if the target
	// is a Kustomization.
	isKustomization, err := hasKustomization(args[0])
	if err != nil {
		return err
	}
	// If the argument is a folder and it has a Kustomization
	// file in it, use Kustomize. If not, just fall back to
	// regular apply.
	if isKustomization {
		o.DeleteFlags.FileNameFlags.Kustomize = &args[0]
	} else {
		o.DeleteFlags.FileNameFlags.Filenames = &args
		o.DeleteFlags.FileNameFlags.Recursive = &t
	}

	return nil
}

// hasKustomization checks if the given path points to a folder
// that contains a Kustomization file. If so, it returns true. Otherwise
// it will return false.
func hasKustomization(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return false, nil
	}

	fileNames := []string{"kustomization.yaml", "kustomization.yml", "Kustomization"}
	var firstErr error
	for _, fileName := range fileNames {
		p := filepath.Join(path, fileName)
		_, err := os.Stat(p)
		if err == nil {
			return true, nil
		}
		if os.IsNotExist(err) {
			continue
		}
		if firstErr != nil {
			firstErr = err
		}
	}
	return false, firstErr
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
