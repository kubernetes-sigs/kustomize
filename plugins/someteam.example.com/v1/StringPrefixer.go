// +build plugin

// Assuming GOPATH is something like
//   ~/gopath
// and this source file is located at
//   $GOPATH/src/sigs.k8s.io/kustomize/plugins/StringPrefixer.go,
// build it like this:
//   dir=$GOPATH/src/sigs.k8s.io/kustomize/plugins
//   go build -buildmode plugin -tags=plugin \
//     -o $dir/StringPrefixer.so $dir/StringPrefixer.go

package main

import (
	"fmt"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

type plugin struct {
	prefix string
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) error {
	name := k.GetName()
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	p.prefix = name + "-"
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	tr, err := transformers.NewNamePrefixSuffixTransformer(
		p.prefix, "",
		config.MakeDefaultConfig().NamePrefix)
	if err != nil {
		return err
	}
	return tr.Transform(m)
}
