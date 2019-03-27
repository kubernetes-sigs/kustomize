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
)

// Literals takes a list of literals.
// Each literal source should be a key and literal value,
// e.g. `somekey=somevalue`
type Literals struct{}

// Get implements the interface for kv plugins.
func (p Literals) Get(root string, args []string) (map[string]string, error) {
	kvs := make(map[string]string)
	for _, s := range args {
		k, v, err := kv.ParseLiteralSource(s)
		if err != nil {
			return nil, err
		}
		kvs[k] = v
	}
	return kvs, nil
}
