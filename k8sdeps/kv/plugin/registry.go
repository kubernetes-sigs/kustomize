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

	"sigs.k8s.io/kustomize/pkg/ifc"
)

// Registry holds all the plugin factories.
type Registry struct {
	factories map[string]Factory
	ldr       ifc.Loader
}

// NewRegistry returns a new Registry loaded with all the factories.
func NewRegistry(ldr ifc.Loader) Registry {
	return Registry{
		ldr: ldr,
		factories: map[string]Factory{
			"go":       newGoFactory(),
			"testonly": newTestonlyFactory(),
		},
	}
}

// Load returns a plugin by type and name,
func (r *Registry) Load(pluginType, name string) (KVSource, error) {
	factory, exists := r.factories[pluginType]
	if !exists {
		return nil, fmt.Errorf("%s is not a valid plugin type", pluginType)
	}
	return factory.load(name)
}

// Root returns the root of the plugins kustomization file.
func (r *Registry) Root() string {
	return r.ldr.Root()
}
