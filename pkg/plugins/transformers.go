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
	kplugin "sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
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

type transformerLoader struct {
	pc  *types.PluginConfig
	ldr ifc.Loader
	rf  *resmap.Factory
}

func NewTransformerLoader(
	pc *types.PluginConfig,
	ldr ifc.Loader, rf *resmap.Factory) transformerLoader {
	return transformerLoader{pc: pc, ldr: ldr, rf: rf}
}

func (l transformerLoader) Load(
	rm resmap.ResMap) ([]transformers.Transformer, error) {
	if len(rm) == 0 {
		return nil, nil
	}
	if !l.pc.GoEnabled {
		return nil, fmt.Errorf("plugins not enabled")
	}
	var result []transformers.Transformer
	for id, res := range rm {
		fileName := pluginFileName(l.pc, id)
		c, err := loadAndConfigurePlugin(fileName, l.ldr, l.rf, res)
		if err != nil {
			return nil, err
		}
		t, ok := c.(transformers.Transformer)
		if !ok {
			return nil, fmt.Errorf("plugin %s not a transformer", fileName)
		}
		result = append(result, t)
	}
	return result, nil
}

func pluginFileName(pc *types.PluginConfig, id resid.ResId) string {
	return filepath.Join(
		pc.DirectoryPath,
		id.Gvk().Group, id.Gvk().Version, id.Gvk().Kind+".so")
}

func loadAndConfigurePlugin(
	fileName string, ldr ifc.Loader,
	rf *resmap.Factory, res *resource.Resource) (Configurable, error) {
	goPlugin, err := plugin.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s fails to load", fileName)
	}
	symbol, err := goPlugin.Lookup(kplugin.PluginSymbol)
	if err != nil {
		return nil, errors.Wrapf(
			err, "plugin %s doesn't have symbol %s",
			fileName, kplugin.PluginSymbol)
	}
	c, ok := symbol.(Configurable)
	if !ok {
		return nil, fmt.Errorf("plugin %s not configurable", fileName)
	}
	err = c.Config(ldr, rf, res)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s fails configuration", fileName)
	}
	return c, nil
}
