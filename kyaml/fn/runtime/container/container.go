// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"time"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/docker"
	runtimeexec "sigs.k8s.io/kustomize/kyaml/fn/runtime/exec"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/kubectl"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Filter struct {
	runtimeutil.ContainerSpec `json:",inline" yaml:",inline"`

	UIDGID string

	UseKubectl        bool
	KubectlGlobalArgs []string
	PodTemplateName   string
	PodStartTimeout   *time.Duration

	Exec runtimeexec.Filter
}

func NewContainer(spec runtimeutil.ContainerSpec, uidgid string, useKubectl bool, podTemplateName string, kubectlGlobalArgs []string, podStartTimeout *time.Duration) Filter {
	f := Filter{ContainerSpec: spec, UIDGID: uidgid, UseKubectl: useKubectl, KubectlGlobalArgs: kubectlGlobalArgs, PodTemplateName: podTemplateName, PodStartTimeout: podStartTimeout}

	return f
}

func (c *Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {

	var ifc kio.Filter
	if c.UseKubectl {
		flt := kubectl.NewContainer(c.ContainerSpec, c.PodTemplateName, c.KubectlGlobalArgs, c.PodStartTimeout)
		flt.Exec.FunctionConfig = c.Exec.FunctionConfig
		flt.Exec.GlobalScope = c.Exec.GlobalScope
		flt.Exec.ResultsFile = c.Exec.ResultsFile
		flt.Exec.DeferFailure = c.Exec.DeferFailure
		ifc = &flt
	} else {
		flt := docker.NewContainer(c.ContainerSpec, c.UIDGID)
		flt.Exec.FunctionConfig = c.Exec.FunctionConfig
		flt.Exec.GlobalScope = c.Exec.GlobalScope
		flt.Exec.ResultsFile = c.Exec.ResultsFile
		flt.Exec.DeferFailure = c.Exec.DeferFailure
		ifc = &flt
	}
	return ifc.Filter(nodes)
}

func (c *Filter) String() string {
	if c.UseKubectl {
		flt := kubectl.NewContainer(c.ContainerSpec, c.PodTemplateName, c.KubectlGlobalArgs, c.PodStartTimeout)
		flt.Exec.FunctionConfig = c.Exec.FunctionConfig
		flt.Exec.GlobalScope = c.Exec.GlobalScope
		flt.Exec.ResultsFile = c.Exec.ResultsFile
		flt.Exec.DeferFailure = c.Exec.DeferFailure
		return flt.String()
	} else {
		flt := docker.NewContainer(c.ContainerSpec, c.UIDGID)
		flt.Exec.FunctionConfig = c.Exec.FunctionConfig
		flt.Exec.GlobalScope = c.Exec.GlobalScope
		flt.Exec.ResultsFile = c.Exec.ResultsFile
		flt.Exec.DeferFailure = c.Exec.DeferFailure
		return flt.String()
	}
}
