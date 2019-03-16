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

package plugin

import (
	"fmt"
	"os"
	"plugin"
)

var _ Factory = &goFactory{}

const (
	dir = "$HOME/.config/kustomize/plugins/kvsource"
)

func newGoFactory() *goFactory {
	return &goFactory{
		plugins: make(map[string]KVSource),
	}
}

type goFactory struct {
	plugins map[string]KVSource
}

func (p *goFactory) load(name string) (KVSource, error) {
	if plug, ok := p.plugins[name]; ok {
		return plug, nil
	}

	goPlugin, err := plugin.Open(fmt.Sprintf("%s/kustomize-%s.so", os.ExpandEnv(dir), name))
	if err != nil {
		return nil, err
	}

	symbol, err := goPlugin.Lookup("Plugin")
	if err != nil {
		return nil, err
	}

	plug, ok := symbol.(KVSource)
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	p.plugins[name] = plug
	return plug, nil
}
