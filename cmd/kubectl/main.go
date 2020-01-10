// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"sigs.k8s.io/kustomize/cmd/kubectl/kubectlcobra"
	"sigs.k8s.io/kustomize/kyaml/commandutil"

	// This is here rather than in the libraries because of
	// https://github.com/kubernetes-sigs/kustomize/issues/2060
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	// enable the config commands
	os.Setenv(commandutil.EnableAlphaCommmandsEnvName, "true")
	if err := kubectlcobra.GetCommand(nil).Execute(); err != nil {
		os.Exit(1)
	}
}
