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
	ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) (err error) {
	p.ldr = ldr
	p.rf = rf
	p.args.GeneratorArgs, err = resmap.GeneratorArgsFromKunstruct(k)
	if err != nil {
		return
	}
	p.args.Type, err = k.GetFieldValue("type")
	if !resmap.IsAcceptableError(err) {
		return
	}
	// panic("hello")
	return nil
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	argsList := make([]types.SecretArgs, 1)
	argsList[0] = p.args
	// panic("ello")
	return p.rf.NewResMapFromSecretArgs(p.ldr, &p.options, argsList)
}
