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
	"log"
	"sort"

	"github.com/ghodss/yaml"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/transformerconfig/defaultconfig"
)

type rpcSlice []ReferencePathConfig

func (s rpcSlice) Len() int      { return len(s) }
func (s rpcSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s rpcSlice) Less(i, j int) bool {
	return s[i].Gvk.IsLessThan(s[j].Gvk)
}

type pcSlice []PathConfig

func (s pcSlice) Len() int      { return len(s) }
func (s pcSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s pcSlice) Less(i, j int) bool {
	return s[i].Gvk.IsLessThan(s[j].Gvk)
}

// TransformerConfig represents the path configurations for different transformations
type TransformerConfig struct {
	NamePrefix        pcSlice  `json:"namePrefix,omitempty" yaml:"namePrefix,omitempty"`
	NameSpace         pcSlice  `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	CommonLabels      pcSlice  `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`
	CommonAnnotations pcSlice  `json:"commonAnnotations,omitempty" yaml:"commonAnnotations,omitempty"`
	NameReference     rpcSlice `json:"nameReference,omitempty" yaml:"nameReference,omitempty"`
	VarReference      pcSlice  `json:"varReference,omitempty" yaml:"varReference,omitempty"`
}

// sortFields provides determinism in logging, tests, etc.
func (t *TransformerConfig) sortFields() {
	sort.Sort(t.NamePrefix)
	sort.Sort(t.NameSpace)
	sort.Sort(t.CommonLabels)
	sort.Sort(t.CommonAnnotations)
	sort.Sort(t.NameReference)
	sort.Sort(t.VarReference)
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
	merged.sortFields()
	return merged
}

// MakeTransformerConfigFromFiles returns a TranformerConfig object from a list of files
func MakeTransformerConfigFromFiles(ldr ifc.Loader, paths []string) (*TransformerConfig, error) {
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
	t.sortFields()
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
