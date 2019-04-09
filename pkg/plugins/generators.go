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

	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/types"
)

type generatorLoader struct {
	pc  *types.PluginConfig
	ldr ifc.Loader
	rf  *resmap.Factory
}

func NewGeneratorLoader(
	pc *types.PluginConfig,
	ldr ifc.Loader, rf *resmap.Factory) generatorLoader {
	return generatorLoader{pc: pc, ldr: ldr, rf: rf}
}

func (l generatorLoader) Load(
	rm resmap.ResMap) ([]transformers.Generator, error) {
	if len(rm) == 0 {
		return nil, nil
	}
	if !l.pc.GoEnabled {
		return nil, fmt.Errorf("plugins not enabled")
	}
	var result []transformers.Generator
	for id, res := range rm {
		c, err := loadAndConfigurePlugin(l.pc.DirectoryPath, id, l.ldr, l.rf, res)
		if err != nil {
			return nil, err
		}
		g, ok := c.(transformers.Generator)
		if !ok {
			return nil, fmt.Errorf("plugin %s not a generator", id.String())
		}
		result = append(result, g)
	}
	return result, nil
}
