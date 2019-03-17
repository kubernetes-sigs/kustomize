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

	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin/builtin"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

var _ Factory = &builtinFactory{}

type builtinFactory struct {
	plugins map[string]KVSource
}

func newBuiltinFactory(ldr ifc.Loader) *builtinFactory {
	return &builtinFactory{
		plugins: map[string]KVSource{
			"literals": builtin.Literals{},
		},
	}
}

func (p *builtinFactory) load(name string) (KVSource, error) {
	if plug, ok := p.plugins[name]; ok {
		return plug, nil
	}
	return nil, fmt.Errorf("plugin %s not found", name)
}
