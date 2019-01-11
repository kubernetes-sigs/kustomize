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

package config

import (
	"log"

	"github.com/ghodss/yaml"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

// Factory makes instances of TransformerConfig.
type Factory struct {
	ldr ifc.Loader
}

// TODO(#606): Setting this to false satisfies the feature
// request in 606.  The todo is to delete the non-active
// code path in a subsequent PR.
const demandExplicitConfig = false

func MakeTransformerConfig(
	ldr ifc.Loader, paths []string) (*TransformerConfig, error) {
	if demandExplicitConfig {
		return loadConfigFromDiskOrDefaults(ldr, paths)
	}
	return mergeCustomConfigWithDefaults(ldr, paths)
}

// loadConfigFromDiskOrDefaults returns a TransformerConfig object
// built from either files or the hardcoded default configs.
// There's no merging, it's one or the other.  This is preferred
// if one wants all configuration to be explicit in version
// control, as opposed to relying on a mix of files and
// hard-coded config.
func loadConfigFromDiskOrDefaults(
	ldr ifc.Loader, paths []string) (*TransformerConfig, error) {
	if paths == nil || len(paths) == 0 {
		return MakeDefaultConfig(), nil
	}
	return NewFactory(ldr).FromFiles(paths)
}

// mergeCustomConfigWithDefaults returns a merger of custom config,
// if any, with default config.
func mergeCustomConfigWithDefaults(
	ldr ifc.Loader, paths []string) (*TransformerConfig, error) {
	t1 := MakeDefaultConfig()
	if len(paths) == 0 {
		return t1, nil
	}
	t2, err := NewFactory(ldr).FromFiles(paths)
	if err != nil {
		return nil, err
	}
	return t1.Merge(t2)
}

func NewFactory(l ifc.Loader) *Factory {
	return &Factory{ldr: l}
}

func (tf *Factory) loader() ifc.Loader {
	if tf.ldr.(ifc.Loader) == nil {
		log.Fatal("no loader")
	}
	return tf.ldr
}

// FromFiles returns a TranformerConfig object from a list of files
func (tf *Factory) FromFiles(
	paths []string) (*TransformerConfig, error) {
	result := &TransformerConfig{}
	for _, path := range paths {
		data, err := tf.loader().Load(path)
		if err != nil {
			return nil, err
		}
		t, err := makeTransformerConfigFromBytes(data)
		if err != nil {
			return nil, err
		}
		result, err = result.Merge(t)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// makeTransformerConfigFromBytes returns a TransformerConfig object from bytes
func makeTransformerConfigFromBytes(data []byte) (*TransformerConfig, error) {
	var t TransformerConfig
	err := yaml.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}
	t.sortFields()
	return &t, nil
}
