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

package builtin

import (
	"sigs.k8s.io/kustomize/k8sdeps/kv"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

// EnvFiles format should be a path to a file to read lines of key=val
// pairs to create a configmap.
// i.e. a Docker .env file or a .ini file.
type EnvFiles struct {
	Ldr ifc.Loader
}

// Get implements the interface for kv plugins.
func (p EnvFiles) Get(root string, args []string) (map[string]string, error) {
	all := make(map[string]string)
	for _, path := range args {
		if path == "" {
			return nil, nil
		}
		content, err := p.Ldr.Load(path)
		if err != nil {
			return nil, err
		}
		kvs, err := kv.KeyValuesFromLines(content)
		if err != nil {
			return nil, err
		}
		for _, pair := range kvs {
			all[pair.Key] = pair.Value
		}
	}
	return all, nil
}
