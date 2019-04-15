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
	configs := getGroupedConfigs(rm)
	for group, res := range configs {
		c, err := loadAndConfigurePlugin(l.pc.DirectoryPath, group, l.ldr, l.rf, res)
		if err != nil {
			return nil, err
		}
		t, ok := c.(transformers.Transformer)
		if !ok {
			return nil, fmt.Errorf("plugin %s not a transformer", group)
		}
		result = append(result, t)
	}
	return result, nil
}
