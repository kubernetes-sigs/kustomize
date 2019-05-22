/*
Copyright 2018 The Kubernetes Authors.

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

// Package configmapandsecret generates configmaps and secrets per generator rules.
package configmapandsecret

import (
	"fmt"
	"unicode/utf8"

	"k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/pkg/types"
)

func makeFreshConfigMap(
	args *types.ConfigMapArgs) *v1.ConfigMap {
	cm := &v1.ConfigMap{}
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	cm.Name = args.Name
	cm.Namespace = args.Namespace
	cm.Data = map[string]string{}
	return cm
}

// MakeConfigMap returns a new ConfigMap, or nil and an error.
func (f *Factory) MakeConfigMap(
	args *types.ConfigMapArgs) (*v1.ConfigMap, error) {
	all, err := f.loadKvPairs(args.GeneratorArgs)
	if err != nil {
		return nil, err
	}
	cm := makeFreshConfigMap(args)
	for _, p := range all {
		err = addKvToConfigMap(cm, p.Key, p.Value)
		if err != nil {
			return nil, err
		}
	}
	if f.options != nil {
		cm.SetLabels(f.options.Labels)
		cm.SetAnnotations(f.options.Annotations)
	}
	return cm, nil
}

// addKvToConfigMap adds the given key and data to the given config map.
// Error if key invalid, or already exists.
func addKvToConfigMap(configMap *v1.ConfigMap, keyName, data string) error {
	if err := errIfInvalidKey(keyName); err != nil {
		return err
	}
	// If the configmap data contains byte sequences that are all in the UTF-8
	// range, we will write it to .Data
	if utf8.Valid([]byte(data)) {
		if _, entryExists := configMap.Data[keyName]; entryExists {
			return fmt.Errorf(keyExistsErrorMsg, keyName, configMap.Data)
		}
		configMap.Data[keyName] = data
		return nil
	}
	// otherwise, it's BinaryData
	if configMap.BinaryData == nil {
		configMap.BinaryData = map[string][]byte{}
	}
	if _, entryExists := configMap.BinaryData[keyName]; entryExists {
		return fmt.Errorf(keyExistsErrorMsg, keyName, configMap.BinaryData)
	}
	configMap.BinaryData[keyName] = []byte(data)
	return nil
}
