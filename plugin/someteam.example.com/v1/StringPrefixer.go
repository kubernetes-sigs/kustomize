// +build plugin

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
