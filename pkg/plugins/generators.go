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
)

func (l *Loader) LoadGenerators(
	ldr ifc.Loader, rm resmap.ResMap) ([]transformers.Generator, error) {
	var result []transformers.Generator
	for _, res := range rm {
		c, err := l.loadAndConfigurePlugin(ldr, res)
		if err != nil {
			return nil, err
		}
		g, ok := c.(transformers.Generator)
		if !ok {
			return nil, fmt.Errorf("plugin %s not a generator", res.Id())
		}
		result = append(result, g)
	}
	return result, nil
}
