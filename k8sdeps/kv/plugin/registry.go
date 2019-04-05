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

	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/types"
)

// Registry holds all the plugin factories.
type Registry struct {
	factories map[types.PluginType]Factory
	ldr       ifc.Loader
}

const (
	TransformerSymbol = "Transformer"
	PluginsDir        = "plugins"
	pluginTypeGo      = types.PluginType("go")
	pluginTypeBuiltIn = types.PluginType("builtin")
)

func ActivePluginConfig() *types.PluginConfig {
	pc := DefaultPluginConfig()
	pc.GoEnabled = true
	return pc
}

func DefaultPluginConfig() *types.PluginConfig {
	return &types.PluginConfig{
		GoEnabled: false,
		DirectoryPath: filepath.Join(
			pgmconfig.ConfigRoot(), PluginsDir),
	}
}

// NewConfiguredRegistry returns a new Registry loaded
// with all the factories and a custom PluginConfig.
func NewConfiguredRegistry(
	ldr ifc.Loader, pc *types.PluginConfig) Registry {
	return Registry{
		ldr: ldr,
		factories: map[types.PluginType]Factory{
			pluginTypeGo:      newGoFactory(pc),
			pluginTypeBuiltIn: newBuiltinFactory(ldr),
		},
	}
}

// NewRegistry returns a new Registry with default config.
func NewRegistry(ldr ifc.Loader) Registry {
	return NewConfiguredRegistry(ldr, &types.PluginConfig{})
}

// Load returns a plugin by type and name.
func (r *Registry) Load(
	pt types.PluginType, name string) (KVSource, error) {
	if pt.IsUndefined() {
		pt = pluginTypeBuiltIn
	}
	factory, exists := r.factories[pt]
	if !exists {
		return nil, fmt.Errorf("%s is not a valid plugin type", pt)
	}
	return factory.load(name)
}

// Root returns the root of the plugins kustomization file.
func (r *Registry) Root() string {
	return r.ldr.Root()
}
