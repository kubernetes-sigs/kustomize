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

// Package transformerconfig provides the functions to load default or user provided configurations
// for different transformers
package transformerconfig

import (
	"github.com/ghodss/yaml"
	"log"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/transformerconfig/defaultconfig"
)

// TransformerConfig represents the path configurations for different transformations
type TransformerConfig struct {
	NamePrefix        []PathConfig          `json:"namePrefix,omitempty" yaml:"namePrefix,omitempty"`
	NameSpace         []PathConfig          `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	CommonLabels      []PathConfig          `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`
	CommonAnnotations []PathConfig          `json:"commonAnnotations,omitempty" yaml:"commonAnnotations,omitempty"`
	NameReference     []ReferencePathConfig `json:"nameReference,omitempty" yaml:"nameReference,omitempty"`
	VarReference      []PathConfig          `json:"varReference,omitempty" yaml:"varReference,omitempty"`
}

// AddPrefixPathConfig adds a PathConfig to NamePrefix
func (t *TransformerConfig) AddPrefixPathConfig(config PathConfig) {
	t.NamePrefix = append(t.NamePrefix, config)
}

// AddLabelPathConfig adds a PathConfig to CommonLabels
func (t *TransformerConfig) AddLabelPathConfig(config PathConfig) {
	t.CommonLabels = append(t.CommonLabels, config)
}

// AddAnnotationPathConfig adds a PathConfig to CommonAnnotations
func (t *TransformerConfig) AddAnnotationPathConfig(config PathConfig) {
	t.CommonAnnotations = append(t.CommonAnnotations, config)
}

// AddNamereferencePathConfig adds a ReferencePathConfig to NameReference
func (t *TransformerConfig) AddNamereferencePathConfig(config ReferencePathConfig) {
	t.NameReference = mergeNameReferencePathConfigs(t.NameReference, []ReferencePathConfig{config})
}

// Merge merges two TransformerConfigs objects into a new TransformerConfig object
func (t *TransformerConfig) Merge(input *TransformerConfig) *TransformerConfig {
	merged := &TransformerConfig{}
	merged.NamePrefix = append(t.NamePrefix, input.NamePrefix...)
	merged.NameSpace = append(t.NameSpace, input.NameSpace...)
	merged.CommonAnnotations = append(t.CommonAnnotations, input.CommonAnnotations...)
	merged.CommonLabels = append(t.CommonLabels, input.CommonLabels...)
	merged.VarReference = append(t.VarReference, input.VarReference...)
	merged.NameReference = mergeNameReferencePathConfigs(t.NameReference, input.NameReference)
	return merged
}

// MakeTransformerConfigFromFiles returns a TranformerConfig object from a list of files
func MakeTransformerConfigFromFiles(ldr loader.Loader, paths []string) (*TransformerConfig, error) {
	result := &TransformerConfig{}
	for _, path := range paths {
		data, err := ldr.Load(path)
		if err != nil {
			return nil, err
		}
		t, err := MakeTransformerConfigFromBytes(data)
		if err != nil {
			return nil, err
		}
		result = result.Merge(t)
	}
	return result, nil
}

// MakeTransformerConfigFromBytes returns a TransformerConfig object from bytes
func MakeTransformerConfigFromBytes(data []byte) (*TransformerConfig, error) {
	var t TransformerConfig
	err := yaml.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// MakeEmptyTransformerConfig returns an empty TransformerConfig object
func MakeEmptyTransformerConfig() *TransformerConfig {
	return &TransformerConfig{}
}

// MakeDefaultTransformerConfig returns a default TransformerConfig.
// This should never fail, hence the Fatal panic.
func MakeDefaultTransformerConfig() *TransformerConfig {
	c, err := MakeTransformerConfigFromBytes(defaultconfig.GetDefaultPathConfigs())
	if err != nil {
		log.Fatalf("Unable to make default transformconfig: %v", err)
	}
	return c
}
