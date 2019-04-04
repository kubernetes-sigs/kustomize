// +build plugin

package main

import (
	"time"

	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

type plugin struct{}

var Transformer plugin

func (p *plugin) Config(k ifc.Kunstructured) error {
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	tr, err := transformers.NewNamePrefixSuffixTransformer(
		time.Now().Format("2006-01-02")+"-", "",
		config.MakeDefaultConfig().NamePrefix)
	if err != nil {
		return err
	}
	return tr.Transform(m)
}
