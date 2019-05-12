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

package plugins

import (
	"fmt"
	"path/filepath"
	"plugin"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/types"
)

type Configurable interface {
	Config(ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) error
}

type Loader struct {
	pc *types.PluginConfig
	rf *resmap.Factory
}

func NewLoader(
	pc *types.PluginConfig, rf *resmap.Factory) *Loader {
	return &Loader{pc: pc, rf: rf}
}

func (l *Loader) LoadGenerators(
	ldr ifc.Loader, rm resmap.ResMap) ([]transformers.Generator, error) {
	var result []transformers.Generator
	for _, res := range rm {
		g, err := l.LoadGenerator(ldr, res)
		if err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, nil
}

func (l *Loader) LoadGenerator(
	ldr ifc.Loader, res *resource.Resource) (transformers.Generator, error) {
	c, err := l.loadAndConfigurePlugin(ldr, res)
	if err != nil {
		return nil, err
	}
	g, ok := c.(transformers.Generator)
	if !ok {
		return nil, fmt.Errorf("plugin %s not a generator", res.Id())
	}
	return g, nil
}

func (l *Loader) LoadTransformers(
	ldr ifc.Loader, rm resmap.ResMap) ([]transformers.Transformer, error) {
	var result []transformers.Transformer
	for _, res := range rm {
		t, err := l.LoadTransformer(ldr, res)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func (l *Loader) LoadTransformer(
	ldr ifc.Loader, res *resource.Resource) (transformers.Transformer, error) {
	c, err := l.loadAndConfigurePlugin(ldr, res)
	if err != nil {
		return nil, err
	}
	t, ok := c.(transformers.Transformer)
	if !ok {
		return nil, fmt.Errorf("plugin %s not a transformer", res.Id())
	}
	return t, nil
}

func pluginPath(id resid.ResId) string {
	return filepath.Join(id.Gvk().Group, id.Gvk().Version, id.Gvk().Kind)
}

func (l *Loader) loadAndConfigurePlugin(
	ldr ifc.Loader, res *resource.Resource) (c Configurable, err error) {
	if !l.pc.GoEnabled {
		return nil, errors.Errorf(
			"plugins not enabled, but trying to load %s", res.Id())
	}
	if p := NewExecPlugin(l.pc.DirectoryPath, res.Id()); p.isAvailable() {
		c = p
	} else {
		c, err = l.loadGoPlugin(res.Id())
		if err != nil {
			return nil, err
		}
	}
	err = c.Config(ldr, l.rf, res)
	if err != nil {
		return nil, errors.Wrapf(
			err, "plugin %s fails configuration", res.Id())
	}
	return c, nil
}

// registry is a means to avoid trying to load the same .so file
// into memory more than once, which results in an error.
// Each test makes its own loader, and tries to load its own plugins,
// but the loaded .so files are in shared memory, so one will get
// "this plugin already loaded" errors if the registry is maintained
// as a Loader instance variable.  So make it a package variable.
var registry = make(map[string]Configurable)

func (l *Loader) loadGoPlugin(id resid.ResId) (c Configurable, err error) {
	var ok bool
	path := pluginPath(id)
	if c, ok = registry[path]; ok {
		return c, nil
	}
	name := filepath.Join(l.pc.DirectoryPath, path)
	p, err := plugin.Open(name + ".so")
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s fails to load", name)
	}
	symbol, err := p.Lookup(PluginSymbol)
	if err != nil {
		return nil, errors.Wrapf(
			err, "plugin %s doesn't have symbol %s",
			name, PluginSymbol)
	}
	c, ok = symbol.(Configurable)
	if !ok {
		return nil, fmt.Errorf("plugin %s not configurable", name)
	}
	registry[path] = c
	return
}
