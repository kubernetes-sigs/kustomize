// +build plugin

package main

import (
	"fmt"

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
	ldr ifc.Loader, rf *resmap.Factory, name string, k []ifc.Kunstructured) error {
	p.ldr = ldr
	p.rf = rf


	var err error
	// TODO: Should validate this.
	p.args.Behavior, err = k[0].GetFieldValue("behavior")
	if err != nil {
		return err
	}

	envFiles, err := k[0].GetStringSlice("envFiles")
	if err != nil {
		return err
	}
	if len(envFiles) > 2 {
		// TODO: refactor to allow this
		return fmt.Errorf("cannot yet accept more than one envFile")
	}
	if len(envFiles) > 0 {
		p.args.EnvSource = envFiles[0]
	}

	p.args.FileSources, err = k[0].GetStringSlice("valueFiles")
	if err != nil {
		return err
	}
	p.args.LiteralSources, err = k[0].GetStringSlice("literals")
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
