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
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/hash"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ConfigMapFactory makes ConfigMaps.
type ConfigMapFactory struct {
	args *types.ConfigMapArgs
	fSys fs.FileSystem
}

// NewConfigMapFactory returns a new ConfigMapFactory.
func NewConfigMapFactory(args *types.ConfigMapArgs, fSys fs.FileSystem) *ConfigMapFactory {
	return &ConfigMapFactory{args: args, fSys: fSys}
}

// MakeUnstructAndGenerateName returns an configmap and the name appended with a hash.
func (f *ConfigMapFactory) MakeUnstructAndGenerateName() (*unstructured.Unstructured, string, error) {
	cm, err := f.MakeConfigMap()
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

// MakeConfigMap returns a new ConfigMap, or nil and an error.
func (f *ConfigMapFactory) MakeConfigMap() (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	cm.Name = f.args.Name
	cm.Data = map[string]string{}

	if f.args.EnvSource != "" {
		if err := f.handleConfigMapFromEnvFileSource(cm); err != nil {
			return nil, err
		}
	}
	if f.args.FileSources != nil {
		if err := f.handleConfigMapFromFileSources(cm); err != nil {
			return nil, err
		}
	}
	if f.args.LiteralSources != nil {
		if err := f.handleConfigMapFromLiteralSources(cm); err != nil {
			return nil, err
		}
	}

	return cm, nil
}

// handleConfigMapFromLiteralSources adds the specified literal source
// information into the provided configMap.
func (f *ConfigMapFactory) handleConfigMapFromLiteralSources(configMap *v1.ConfigMap) error {
	for _, literalSource := range f.args.LiteralSources {
		keyName, value, err := ParseLiteralSource(literalSource)
		if err != nil {
			return err
		}
		err = addKeyFromLiteralToConfigMap(configMap, keyName, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// handleConfigMapFromFileSources adds the specified file source information
// into the provided configMap
func (f *ConfigMapFactory) handleConfigMapFromFileSources(configMap *v1.ConfigMap) error {
	for _, fileSource := range f.args.FileSources {
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

// HandleConfigMapFromEnvFileSource adds the specified env file source information
// into the provided configMap
func (f *ConfigMapFactory) handleConfigMapFromEnvFileSource(configMap *v1.ConfigMap) error {
	if !f.fSys.Exists(f.args.EnvSource) {
		return fmt.Errorf("unable to read configmap env file %s", f.args.EnvSource)
	}
	if f.fSys.IsDir(f.args.EnvSource) {
		return fmt.Errorf("env config file %s cannot be a directory", f.args.EnvSource)
	}

	return addFromEnvFile(f.args.EnvSource, func(key, value string) error {
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
