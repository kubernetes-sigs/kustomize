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

// Files is a list of file sources.
// Each file source can be specified using its file path, in which case file
// basename will be used as configmap key, or optionally with a key and file
// path, in which case the given key will be used.
// Specifying a directory will iterate each named file in the directory
// whose basename is a valid configmap key.
type Files struct {
	Ldr ifc.Loader
}

// Get implements the interface for kv plugins.
func (p Files) Get(root string, args []string) ([]kv.Pair, error) {
	var kvs []kv.Pair
	for _, s := range args {
		k, fPath, err := kv.ParseFileSource(s)
		if err != nil {
			return nil, err
		}
		content, err := p.Ldr.Load(fPath)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv.Pair{Key: k, Value: string(content)})
	}
	return kvs, nil
}
