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
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
)

const generatorSymbol = "Generator"

type Generatable interface {
	Generate() (resmap.ResMap, error)
}

type generatorLoader struct {
	pluginDir string
	enabled   bool
	rf        *resmap.Factory
}

func NewGeneratorLoader(b bool, f *resmap.Factory) generatorLoader {
	return generatorLoader{
		pluginDir: filepath.Join(pgmconfig.ConfigRoot(), pgmconfig.PluginsDir),
		enabled:   b,
		rf:        f,
	}
}

func (l generatorLoader) Load(rm resmap.ResMap) (resmap.ResMap, error) {
	if len(rm) == 0 {
		return nil, nil
	}
	if !l.enabled {
		return nil, fmt.Errorf("plugin is not enabled")
	}
	var result resmap.ResMap
	for id, res := range rm {
		r, err := l.load(id, res)
		if err != nil {
			return nil, err
		}
		result, err = resmap.MergeWithErrorOnIdCollision(result, r)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (l generatorLoader) load(id resid.ResId, res *resource.Resource) (resmap.ResMap, error) {
	fileName := filepath.Join(l.pluginDir, id.Gvk().Kind+".so")
	goPlugin, err := plugin.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("plugin %s file not opened", fileName)
	}

	symbol, err := goPlugin.Lookup(generatorSymbol)
	if err != nil {
		return nil, fmt.Errorf("plugin %s fails lookup", fileName)
	}

	c, ok := symbol.(Configurable)
	if !ok {
		return nil, fmt.Errorf("plugin %s not configurable", fileName)
	}
	err = c.Config(res)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s fails configuration", fileName)
	}

	g, ok := c.(Generatable)
	if !ok {
		return nil, fmt.Errorf("plugin %s not a transformer", fileName)
	}
	return g.Generate()
}
