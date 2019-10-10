// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"fmt"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

type Loader struct {
	pc *types.PluginConfig
	rf *resmap.Factory
}

func NewLoader(
	pc *types.PluginConfig, rf *resmap.Factory) *Loader {
	return &Loader{pc: pc, rf: rf}
}

func (l *Loader) LoadGenerators(
	ldr ifc.Loader, rm resmap.ResMap) ([]resmap.Generator, error) {
	var result []resmap.Generator
	for _, res := range rm.Resources() {
		g, err := l.LoadGenerator(ldr, res)
		if err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, nil
}

func (l *Loader) LoadGenerator(
	ldr ifc.Loader, res *resource.Resource) (resmap.Generator, error) {
	c, err := l.loadAndConfigurePlugin(ldr, res)
	if err != nil {
		return nil, err
	}
	g, ok := c.(resmap.Generator)
	if !ok {
		return nil, fmt.Errorf("plugin %s not a generator", res.OrgId())
	}
	return g, nil
}

func (l *Loader) LoadTransformers(
	ldr ifc.Loader, rm resmap.ResMap) ([]resmap.Transformer, error) {
	var result []resmap.Transformer
	for _, res := range rm.Resources() {
		t, err := l.LoadTransformer(ldr, res)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func (l *Loader) LoadTransformer(
	ldr ifc.Loader, res *resource.Resource) (resmap.Transformer, error) {
	c, err := l.loadAndConfigurePlugin(ldr, res)
	if err != nil {
		return nil, err
	}
	t, ok := c.(resmap.Transformer)
	if !ok {
		return nil, fmt.Errorf("plugin %s not a transformer", res.OrgId())
	}
	return t, nil
}

func relativePluginPath(id resid.ResId) string {
	return filepath.Join(
		id.Group,
		id.Version,
		strings.ToLower(id.Kind))
}

func AbsolutePluginPath(pc *types.PluginConfig, id resid.ResId) string {
	return filepath.Join(
		pc.DirectoryPath, relativePluginPath(id), id.Kind)
}

func (l *Loader) absolutePluginPath(id resid.ResId) string {
	return AbsolutePluginPath(l.pc, id)
}

func isBuiltinPlugin(res *resource.Resource) bool {
	// TODO: the special string should appear in Group, not Version.
	return res.GetGvk().Group == "" &&
		res.GetGvk().Version == BuiltinPluginApiVersion
}

func (l *Loader) loadAndConfigurePlugin(
	ldr ifc.Loader, res *resource.Resource) (c resmap.Configurable, err error) {
	if isBuiltinPlugin(res) {
		// Instead of looking for and loading a .so file, just
		// instantiate the plugin from a generated factory
		// function (see "pluginator").  Being able to do this
		// is what makes a plugin "builtin".
		c, err = l.makeBuiltinPlugin(res.GetGvk())
	} else if l.pc.Enabled {
		c, err = l.loadPlugin(res.OrgId())
	} else {
		err = notEnabledErr(res.OrgId().Kind)
	}
	if err != nil {
		return nil, err
	}
	yaml, err := res.AsYAML()
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling yaml from res %s", res.OrgId())
	}
	err = c.Config(ldr, l.rf, yaml)
	if err != nil {
		return nil, errors.Wrapf(
			err, "plugin %s fails configuration", res.OrgId())
	}
	return c, nil
}

func (l *Loader) makeBuiltinPlugin(r gvk.Gvk) (resmap.Configurable, error) {
	bpt := GetBuiltinPluginType(r.Kind)
	if f, ok := GeneratorFactories[bpt]; ok {
		return f(), nil
	}
	if f, ok := TransformerFactories[bpt]; ok {
		return f(), nil
	}
	return nil, errors.Errorf("unable to load builtin %s", r)
}

func (l *Loader) loadPlugin(resId resid.ResId) (resmap.Configurable, error) {
	p := NewExecPlugin(l.absolutePluginPath(resId))
	if p.isAvailable() {
		return p, nil
	}
	c, err := l.loadGoPlugin(resId)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// registry is a means to avoid trying to load the same .so file
// into memory more than once, which results in an error.
// Each test makes its own loader, and tries to load its own plugins,
// but the loaded .so files are in shared memory, so one will get
// "this plugin already loaded" errors if the registry is maintained
// as a Loader instance variable.  So make it a package variable.
var registry = make(map[string]resmap.Configurable)

func (l *Loader) loadGoPlugin(id resid.ResId) (resmap.Configurable, error) {
	regId := relativePluginPath(id)
	if c, ok := registry[regId]; ok {
		return copyPlugin(c), nil
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
	c, ok := symbol.(resmap.Configurable)
	if !ok {
		return nil, fmt.Errorf("plugin %s not configurable", regId)
	}
	registry[regId] = c
	return copyPlugin(c), nil
}

func copyPlugin(c resmap.Configurable) resmap.Configurable {
	indirect := reflect.Indirect(reflect.ValueOf(c))
	newIndirect := reflect.New(indirect.Type())
	newIndirect.Elem().Set(reflect.ValueOf(indirect.Interface()))
	newNamed := newIndirect.Interface()
	return newNamed.(resmap.Configurable)
}
