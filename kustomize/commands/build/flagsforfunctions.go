// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kustomize/api/types"
)

func AddFunctionFlags(set *pflag.FlagSet, o *types.FnPluginLoadingOptions) {
	set.BoolVar(
		&o.EnableExec, "enable-exec", false, /*do not change!*/
		"enable support for exec functions -- note: exec functions run arbitrary code -- do not use for untrusted configs!!! (Alpha)")
	set.BoolVar(
		&o.EnableStar, "enable-star", false,
		"enable support for starlark functions. (Alpha)")
	set.BoolVar(
		&o.Network, "network", false,
		"enable network access for functions that declare it")
	set.StringVar(
		&o.NetworkName, "network-name", "bridge",
		"the docker network to run the container in")
	set.StringArrayVar(
		&o.Mounts, "mount", []string{},
		"a list of storage options read from the filesystem")
	set.StringArrayVarP(
		&o.Env, "env", "e", []string{},
		"a list of environment variables to be used by functions")
}
