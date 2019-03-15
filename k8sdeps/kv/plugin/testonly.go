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

// testonly is temporary until we have builtin plugins to use in the tests.

package plugin

import (
	"sigs.k8s.io/kustomize/k8sdeps/kv"
)

var _ Factory = &testonlyFactory{}

func newTestonlyFactory() *testonlyFactory {
	return &testonlyFactory{}
}

type testonlyFactory struct{}

func (p testonlyFactory) Get(_ string, args []string) ([]kv.Pair, error) {
	var kvs []kv.Pair
	for _, arg := range args {
		kvs = append(kvs, kv.Pair{Key: "k_" + arg, Value: "v_" + arg})
	}
	return kvs, nil
}

func (p *testonlyFactory) load(name string) (KVSource, error) {
	return p, nil
}
