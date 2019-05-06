// +build plugin

package main

import (
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"
)

type plugin struct {
	ldr     ifc.Loader
	rf      *resmap.Factory
	options types.GeneratorOptions
	args    types.SecretArgs
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) error {
	p.ldr = ldr
	p.rf = rf
	var err error
	// TODO: validate behavior values.
	p.args.Behavior, err = k.GetFieldValue("behavior")
	if err != nil {
		return err
	}
	p.args.EnvSources, err = k.GetStringSlice("envFiles")
	if err != nil {
		return err
	}
	p.args.FileSources, err = k.GetStringSlice("valueFiles")
	if err != nil {
		return err
	}
	p.args.LiteralSources, err = k.GetStringSlice("literals")
	if err != nil {
		return err
	}
	return nil
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	argsList := make([]types.SecretArgs, 1)
	argsList[0] = p.args
	return p.rf.NewResMapFromSecretArgs(p.ldr, &p.options, argsList)
}
