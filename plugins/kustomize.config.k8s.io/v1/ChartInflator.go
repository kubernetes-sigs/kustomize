// +build plugin

package main

import (
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
)

type plugin struct {
	ldr ifc.Loader
	rf  *resmap.Factory
	// Root directory of chart, in which one
	// finds Chart.yaml, values.yaml, templates/, etc.
	root string
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) error {
	var err error
	p.root, err = k.GetFieldValue("root")
	if err != nil {
		return err
	}
	return nil
}

// TODO: vendor in helm code, make same library calls as made by
// "helm template mychart" but return it here rather than stdout.
func (p *plugin) inflateChart() []byte {
	return []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: ThisIsJustAPlaceHolder
`)
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	return p.rf.NewResMapFromBytes(p.inflateChart())
}
