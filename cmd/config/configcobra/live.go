// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configcobra

import (
	"flag"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/cli-utils/cmd/apply"
	"sigs.k8s.io/cli-utils/cmd/destroy"
	"sigs.k8s.io/cli-utils/cmd/diff"
	"sigs.k8s.io/cli-utils/cmd/initcmd"
	"sigs.k8s.io/cli-utils/cmd/preview"
	"sigs.k8s.io/cli-utils/pkg/util/factory"
)

func GetLive(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "live",
		Short: "Commands for reading and writing resources to a cluster.",
	}
	ioStreams := genericclioptions.IOStreams{
		In:     cmd.InOrStdin(),
		Out:    cmd.OutOrStdout(),
		ErrOut: cmd.ErrOrStderr(),
	}

	flags := cmd.PersistentFlags()
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.AddFlags(flags)
	userAgentKubeConfigFlags := &UserAgentKubeConfigFlags{
		Delegate:  kubeConfigFlags,
		UserAgent: "kustomize",
	}
	matchVersionKubeConfigFlags := util.NewMatchVersionFlags(
		&factory.CachingRESTClientGetter{
			Delegate: userAgentKubeConfigFlags,
		},
	)
	matchVersionKubeConfigFlags.AddFlags(cmd.PersistentFlags())
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	f := util.NewFactory(matchVersionKubeConfigFlags)

	applyCmd := apply.ApplyCommand(f, ioStreams)
	_ = applyCmd.Flags().MarkHidden("no-prune")

	cmd.AddCommand(
		applyCmd,
		initcmd.NewCmdInit(f, ioStreams),
		preview.PreviewCommand(f, ioStreams),
		diff.NewCmdDiff(f, ioStreams),
		destroy.DestroyCommand(f, ioStreams))
	return cmd
}

type UserAgentKubeConfigFlags struct {
	Delegate  genericclioptions.RESTClientGetter
	UserAgent string
}

func (u *UserAgentKubeConfigFlags) ToRESTConfig() (*rest.Config, error) {
	clientConfig, err := u.Delegate.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	if u.UserAgent != "" {
		clientConfig.UserAgent = u.UserAgent
	}
	return clientConfig, nil
}

func (u *UserAgentKubeConfigFlags) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return u.Delegate.ToDiscoveryClient()
}

func (u *UserAgentKubeConfigFlags) ToRESTMapper() (meta.RESTMapper, error) {
	return u.Delegate.ToRESTMapper()
}

func (u *UserAgentKubeConfigFlags) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return u.Delegate.ToRawKubeConfigLoader()
}
