// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
)

func AddFunctionBasicsFlags(set *pflag.FlagSet) {
	set.BoolVar(
		&theFlags.fnOptions.Network, "network", false,
		"enable network access for functions that declare it")
	set.StringVar(
		&theFlags.fnOptions.NetworkName, "network-name", "bridge",
		"the docker network to run the container in")
	set.StringArrayVar(
		&theFlags.fnOptions.Mounts, "mount", []string{},
		"a list of storage options read from the filesystem")
	set.StringArrayVarP(
		&theFlags.fnOptions.Env, "env", "e", []string{},
		"a list of environment variables to be used by functions")
	set.BoolVar(
		&theFlags.fnOptions.AsCurrentUser, "as-current-user", false,
		"use the uid and gid of the command executor to run the function in the container")
}

func AddFunctionAlphaEnablementFlags(set *pflag.FlagSet) {
	set.BoolVar(
		&theFlags.fnOptions.EnableExec, "enable-exec", false,
		"enable support for exec functions (raw executables); "+
			"do not use for untrusted configs! (Alpha)")
	set.BoolVar(
		&theFlags.fnOptions.EnableStar, "enable-star", false,
		"enable support for starlark functions. (Alpha)")

	set.BoolVar(
		&theFlags.fnOptions.UseKubectl, "enable-in-a-pod", false,
		"enable support for krm-functions running in pod instead of docker; "+
			"requires kubectl. (Alpha)")
	set.StringVar(
		&theFlags.fnOptions.KubectlGlobalArgs, "pod-kubectl-args", "",
		"list of global arguments to be used to run krm-functions in pod, "+
			"e.g. -n namespace. (Alpha)")

	set.StringVar(
		&theFlags.fnOptions.PodTemplateName, "pod-template-name", "",
		"тame of PodTemplate that will be used for creation of the pod to"+
			"run krm-function. Must exist in the same namespace where"+
			"krm-function is going to be ran (Alpha)")

	set.StringVar(
		&theFlags.fnOptions.PodStartTimeout, "pod-start-timeout", "",
		"Maximum amount of time to wait until the pod start. (Alpha)")
}
