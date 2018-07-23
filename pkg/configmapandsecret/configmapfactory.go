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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/hash"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ConfigMapFactory makes ConfigMaps.
type ConfigMapFactory struct {
	fSys fs.FileSystem
	ldr  loader.Loader
}

// NewConfigMapFactory returns a new ConfigMapFactory.
func NewConfigMapFactory(
	fSys fs.FileSystem, l loader.Loader) *ConfigMapFactory {
	return &ConfigMapFactory{fSys: fSys, ldr: l}
}

// MakeUnstructAndGenerateName returns an configmap and the name appended with a hash.
func (f *ConfigMapFactory) MakeUnstructAndGenerateName(
	args *types.ConfigMapArgs) (*unstructured.Unstructured, string, error) {
	cm, err := f.MakeConfigMap1(args)
	if err != nil {
		return nil, "", err
	}
	h, err := hash.ConfigMapHash(cm)
	if err != nil {
		return nil, "", err
	}
	nameWithHash := fmt.Sprintf("%s-%s", cm.GetName(), h)
	unstructuredCM, err := objectToUnstructured(cm)
	return unstructuredCM, nameWithHash, err
}

func objectToUnstructured(in runtime.Object) (*unstructured.Unstructured, error) {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out unstructured.Unstructured
	err = out.UnmarshalJSON(marshaled)
	return &out, err
}

func (f *ConfigMapFactory) makeFreshConfigMap(
	args *types.ConfigMapArgs) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	cm.Name = args.Name
	cm.Data = map[string]string{}
	return cm
}

// MakeConfigMap1 returns a new ConfigMap, or nil and an error.
func (f *ConfigMapFactory) MakeConfigMap1(
	args *types.ConfigMapArgs) (*corev1.ConfigMap, error) {
	cm := f.makeFreshConfigMap(args)
	if args.EnvSource != "" {
		if err := f.handleConfigMapFromEnvFileSource(cm, args); err != nil {
			return nil, err
		}
	}
	if args.FileSources != nil {
		if err := f.handleConfigMapFromFileSources(cm, args); err != nil {
			return nil, err
		}
	}
	if args.LiteralSources != nil {
		if err := f.handleConfigMapFromLiteralSources(cm, args.LiteralSources); err != nil {
			return nil, err
		}
	}
	return cm, nil
}

// MakeConfigMap2 returns a new ConfigMap, or nil and an error.
// TODO: Get rid of the nearly duplicated code in MakeConfigMap1 vs MakeConfigMap2
func (f *ConfigMapFactory) MakeConfigMap2(
	args *types.ConfigMapArgs) (*corev1.ConfigMap, error) {
	var all []kvPair
	var err error
	cm := f.makeFreshConfigMap(args)

	pairs, err := keyValuesFromEnvFile(f.ldr, args.EnvSource)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"env source file: %s",
			args.EnvSource))
	}
	all = append(all, pairs...)

	pairs, err = keyValuesFromLiteralSources(args.LiteralSources)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"literal sources %v", args.LiteralSources))
	}
	all = append(all, pairs...)

	pairs, err = keyValuesFromFileSources(f.ldr, args.FileSources)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"file sources: %v", args.FileSources))
	}
	all = append(all, pairs...)

	for _, kv := range all {
		err = addKeyFromLiteralToConfigMap(cm, kv.key, kv.value)
		if err != nil {
			return nil, err
		}
	}
	return cm, nil
}

func keyValuesFromLiteralSources(sources []string) ([]kvPair, error) {
	var kvs []kvPair
	for _, s := range sources {
		k, v, err := ParseLiteralSource(s)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kvPair{key: k, value: v})
	}
	return kvs, nil
}

// handleConfigMapFromLiteralSources adds the specified literal source
// information into the provided configMap.
func (f *ConfigMapFactory) handleConfigMapFromLiteralSources(
	configMap *v1.ConfigMap, sources []string) error {
	for _, s := range sources {
		k, v, err := ParseLiteralSource(s)
		if err != nil {
			return err
		}
		err = addKeyFromLiteralToConfigMap(configMap, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func keyValuesFromFileSources(ldr loader.Loader, sources []string) ([]kvPair, error) {
	var kvs []kvPair
	for _, s := range sources {
		k, fPath, err := ParseFileSource(s)
		if err != nil {
			return nil, err
		}
		content, err := ldr.Load(fPath)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kvPair{key: k, value: string(content)})
	}
	return kvs, nil
}

// handleConfigMapFromFileSources adds the specified file source information
// into the provided configMap
func (f *ConfigMapFactory) handleConfigMapFromFileSources(
	configMap *v1.ConfigMap, args *types.ConfigMapArgs) error {
	for _, fileSource := range args.FileSources {
		keyName, filePath, err := ParseFileSource(fileSource)
		if err != nil {
			return err
		}
		if !f.fSys.Exists(filePath) {
			return fmt.Errorf("unable to read configmap source file %s", filePath)
		}
		if f.fSys.IsDir(filePath) {
			if strings.Contains(fileSource, "=") {
				return fmt.Errorf("cannot give a key name for a directory path")
			}
			fileList, err := ioutil.ReadDir(filePath)
			if err != nil {
				return fmt.Errorf("error listing files in %s: %v", filePath, err)
			}
			for _, item := range fileList {
				itemPath := path.Join(filePath, item.Name())
				if item.Mode().IsRegular() {
					keyName = item.Name()
					err = addKeyFromFileToConfigMap(configMap, keyName, itemPath)
					if err != nil {
						return err
					}
				}
			}
		} else {
			if err := addKeyFromFileToConfigMap(configMap, keyName, filePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func keyValuesFromEnvFile(l loader.Loader, path string) ([]kvPair, error) {
	if path == "" {
		return nil, nil
	}
	content, err := l.Load(path)
	if err != nil {
		return nil, err
	}
	return keyValuesFromLines(content)
}

// HandleConfigMapFromEnvFileSource adds the specified env file source information
// into the provided configMap
func (f *ConfigMapFactory) handleConfigMapFromEnvFileSource(
	configMap *v1.ConfigMap, args *types.ConfigMapArgs) error {
	if !f.fSys.Exists(args.EnvSource) {
		return fmt.Errorf("unable to read configmap env file %s", args.EnvSource)
	}
	if f.fSys.IsDir(args.EnvSource) {
		return fmt.Errorf("env config file %s cannot be a directory", args.EnvSource)
	}
	return addFromEnvFile(args.EnvSource, func(key, value string) error {
		return addKeyFromLiteralToConfigMap(configMap, key, value)
	})
}

// addKeyFromFileToConfigMap adds a key with the given name to a ConfigMap, populating
// the value with the content of the given file path, or returns an error.
func addKeyFromFileToConfigMap(configMap *v1.ConfigMap, keyName, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return addKeyFromLiteralToConfigMap(configMap, keyName, string(data))
}

// addKeyFromLiteralToConfigMap adds the given key and data to the given config map,
// returning an error if the key is not valid or if the key already exists.
func addKeyFromLiteralToConfigMap(configMap *v1.ConfigMap, keyName, data string) error {
	// Note, the rules for ConfigMap keys are the exact same as the ones for SecretKeys.
	if errs := validation.IsConfigMapKey(keyName); len(errs) != 0 {
		return fmt.Errorf("%q is not a valid key name for a ConfigMap: %s", keyName, strings.Join(errs, ";"))
	}
	if _, entryExists := configMap.Data[keyName]; entryExists {
		return fmt.Errorf("cannot add key %s, another key by that name already exists: %v", keyName, configMap.Data)
	}
	configMap.Data[keyName] = data
	return nil
}

// ParseFileSource parses the source given.
//
//  Acceptable formats include:
//   1.  source-path: the basename will become the key name
//   2.  source-name=source-path: the source-name will become the key name and
//       source-path is the path to the key file.
//
// Key names cannot include '='.
func ParseFileSource(source string) (keyName, filePath string, err error) {
	numSeparators := strings.Count(source, "=")
	switch {
	case numSeparators == 0:
		return path.Base(source), source, nil
	case numSeparators == 1 && strings.HasPrefix(source, "="):
		return "", "", fmt.Errorf("key name for file path %v missing", strings.TrimPrefix(source, "="))
	case numSeparators == 1 && strings.HasSuffix(source, "="):
		return "", "", fmt.Errorf("file path for key name %v missing", strings.TrimSuffix(source, "="))
	case numSeparators > 1:
		return "", "", errors.New("key names or file paths cannot contain '='")
	default:
		components := strings.Split(source, "=")
		return components[0], components[1], nil
	}
}

// ParseLiteralSource parses the source key=val pair into its component pieces.
// This functionality is distinguished from strings.SplitN(source, "=", 2) since
// it returns an error in the case of empty keys, values, or a missing equals sign.
func ParseLiteralSource(source string) (keyName, value string, err error) {
	// leading equal is invalid
	if strings.Index(source, "=") == 0 {
		return "", "", fmt.Errorf("invalid literal source %v, expected key=value", source)
	}
	// split after the first equal (so values can have the = character)
	items := strings.SplitN(source, "=", 2)
	if len(items) != 2 {
		return "", "", fmt.Errorf("invalid literal source %v, expected key=value", source)
	}

	return items[0], items[1], nil
}
