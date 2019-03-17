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
	"path/filepath"
	"plugin"

	"sigs.k8s.io/kustomize/pkg/types"
)

var _ Factory = &goFactory{}

const (
	kvSourcesDir            = "kvSources"
	EnableGoPluginsFlagName = "enable_alpha_goplugins_accept_panic_risk"
	EnableGoPluginsFlagHelp = `
Warning: the main program may panic and exit on an
attempt to use a goplugin that was compiled under
conditions differing from the those in effect when
main was compiled. It's safest to use this flag in
the context of a container image holding both the
main and the goplugins it needs, all built on the
same machine, with the same transitive libs and
the same compiler version.
`
	errorFmt = `
enable go plugins by specifying flag
  --%s
Place .so files in
  %s
%s
`
)

func newGoFactory(c *types.PluginConfig) *goFactory {
	return &goFactory{
		config:  c,
		plugins: make(map[string]KVSource),
	}
}

type goFactory struct {
	config  *types.PluginConfig
	plugins map[string]KVSource
}

func (p *goFactory) load(name string) (KVSource, error) {
	if plug, ok := p.plugins[name]; ok {
		return plug, nil
	}

	dir := filepath.Join(
		p.config.DirectoryPath,
		kvSourcesDir)
	if !p.config.GoEnabled {
		return nil, fmt.Errorf(
			errorFmt,
			EnableGoPluginsFlagName,
			dir,
			EnableGoPluginsFlagHelp)
	}

	goPlugin, err := plugin.Open(
		filepath.Join(dir, name+".so"))
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
