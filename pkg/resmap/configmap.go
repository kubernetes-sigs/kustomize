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

package resmap

import (
	"fmt"
	"strings"

	cutil "github.com/kubernetes-sigs/kustomize/pkg/configmapandsecret/util"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

func newResourceFromConfigMap(l loader.Loader, cmArgs types.ConfigMapArgs) (*resource.Resource, error) {
	cm, err := makeConfigMap(l, cmArgs)
	if err != nil {
		return nil, err
	}
	if cmArgs.Behavior == "" {
		cmArgs.Behavior = "create"
	}
	return resource.NewResourceWithBehavior(cm, resource.NewGenerationBehavior(cmArgs.Behavior))
}

func makeConfigMap(l loader.Loader, cmArgs types.ConfigMapArgs) (*corev1.ConfigMap, error) {
	var envPairs, literalPairs, filePairs []kvPair
	var err error

	cm := &corev1.ConfigMap{}
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	cm.Name = cmArgs.Name
	cm.Data = map[string]string{}

	if cmArgs.EnvSource != "" {
		envPairs, err = keyValuesFromEnvFile(l, cmArgs.EnvSource)
		if err != nil {
			return nil, fmt.Errorf("error reading keys from env source file: %s %v", cmArgs.EnvSource, err)
		}
	}

	literalPairs, err = keyValuesFromLiteralSources(cmArgs.LiteralSources)
	if err != nil {
		return nil, fmt.Errorf("error reading key values from literal sources: %v", err)
	}

	filePairs, err = keyValuesFromFileSources(l, cmArgs.FileSources)
	if err != nil {
		return nil, fmt.Errorf("error reading key values from file sources: %v", err)
	}

	allPairs := append(append(envPairs, literalPairs...), filePairs...)

	// merge key value pairs from all the sources
	for _, kv := range allPairs {
		err = addKV(cm.Data, kv)
		if err != nil {
			return nil, fmt.Errorf("error adding key in configmap: %v", err)
		}
	}

	return cm, nil
}

func keyValuesFromEnvFile(l loader.Loader, path string) ([]kvPair, error) {
	content, err := l.Load(path)
	if err != nil {
		return nil, err
	}
	return keyValuesFromLines(content)
}

func keyValuesFromLiteralSources(sources []string) ([]kvPair, error) {
	var kvs []kvPair
	for _, s := range sources {
		// TODO: move ParseLiteralSource in this file
		k, v, err := cutil.ParseLiteralSource(s)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kvPair{key: k, value: v})
	}
	return kvs, nil
}

func keyValuesFromFileSources(l loader.Loader, sources []string) ([]kvPair, error) {
	var kvs []kvPair

	for _, s := range sources {
		key, path, err := cutil.ParseFileSource(s)
		if err != nil {
			return nil, err
		}
		fileContent, err := l.Load(path)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kvPair{key: key, value: string(fileContent)})
	}
	return kvs, nil
}

// addKV adds key-value pair to the provided map.
func addKV(m map[string]string, kv kvPair) error {
	if errs := validation.IsConfigMapKey(kv.key); len(errs) != 0 {
		return fmt.Errorf("%q is not a valid key name: %s", kv.key, strings.Join(errs, ";"))
	}
	if _, exists := m[kv.key]; exists {
		return fmt.Errorf("key %s already exists: %v.", kv.key, m)
	}
	m[kv.key] = kv.value
	return nil
}

// NewResMapFromConfigMapArgs returns a Resource slice given a configmap metadata slice from kustomization file.
func NewResMapFromConfigMapArgs(loader loader.Loader, cmList []types.ConfigMapArgs) (ResMap, error) {
	allResources := []*resource.Resource{}
	for _, cm := range cmList {
		res, err := newResourceFromConfigMap(loader, cm)
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, res)
	}
	return newResMapFromResourceSlice(allResources)
}
