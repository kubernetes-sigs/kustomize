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
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/types"
)

type Configurable interface {
	Config(ldr ifc.Loader, rf *resmap.Factory, config []byte) error
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

func relativePluginPath(id resid.ResId) string {
	return filepath.Join(
		id.Gvk().Group,
		id.Gvk().Version,
		strings.ToLower(id.Gvk().Kind))
}

func AbsolutePluginPath(pc *types.PluginConfig, id resid.ResId) string {
	return filepath.Join(
		pc.DirectoryPath, relativePluginPath(id), id.Gvk().Kind)
}

func (l *Loader) absolutePluginPath(id resid.ResId) string {
	return AbsolutePluginPath(l.pc, id)
}

func (l *Loader) loadAndConfigurePlugin(
	ldr ifc.Loader, res *resource.Resource) (c Configurable, err error) {
	if !l.pc.GoEnabled {
		return nil, errors.Errorf(
			"plugins not enabled, but trying to load %s", res.Id())
	}
	if p := NewExecPlugin(
		l.absolutePluginPath(res.Id())); p.isAvailable() {
		c = p
	} else {
		c, err = l.loadGoPlugin(res.Id())
		if err != nil {
			return nil, err
		}
	}
	yaml, err := res.AsYAML()
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling yaml from res %s", res.Id())
	}
	err = c.Config(ldr, l.rf, yaml)
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
	regId := relativePluginPath(id)
	var ok bool
	if c, ok = registry[regId]; ok {
		return c, nil
	}
	absPath := l.absolutePluginPath(id)
	p, err := plugin.Open(absPath + ".so")
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s fails to load", absPath)
	}
	symbol, err := p.Lookup(PluginSymbol)
	if err != nil {
		return nil, errors.Wrapf(
			err, "plugin %s doesn't have symbol %s",
			regId, PluginSymbol)
	}
	c, ok = symbol.(Configurable)
	if !ok {
		return nil, fmt.Errorf("plugin %s not configurable", regId)
	}
	registry[regId] = c
	return
}
