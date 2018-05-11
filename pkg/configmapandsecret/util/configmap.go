/*
Copyright 2016 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// handleConfigMapFromLiteralSources adds the specified literal source
// information into the provided configMap.
func HandleConfigMapFromLiteralSources(configMap *v1.ConfigMap, literalSources []string) error {
	for _, literalSource := range literalSources {
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
func HandleConfigMapFromFileSources(configMap *v1.ConfigMap, fileSources []string) error {
	for _, fileSource := range fileSources {
		keyName, filePath, err := ParseFileSource(fileSource)
		if err != nil {
			return err
		}
		info, err := os.Stat(filePath)
		if err != nil {
			switch err := err.(type) {
			case *os.PathError:
				return fmt.Errorf("error reading %s: %v", filePath, err.Err)
			default:
				return fmt.Errorf("error reading %s: %v", filePath, err)
			}
		}
		if info.IsDir() {
			if strings.Contains(fileSource, "=") {
				return fmt.Errorf("cannot give a key name for a directory path.")
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

// handleConfigMapFromEnvFileSource adds the specified env file source information
// into the provided configMap
func HandleConfigMapFromEnvFileSource(configMap *v1.ConfigMap, envFileSource string) error {
	info, err := os.Stat(envFileSource)
	if err != nil {
		switch err := err.(type) {
		case *os.PathError:
			return fmt.Errorf("error reading %s: %v", envFileSource, err.Err)
		default:
			return fmt.Errorf("error reading %s: %v", envFileSource, err)
		}
	}
	if info.IsDir() {
		return fmt.Errorf("env config file cannot be a directory")
	}

	return addFromEnvFile(envFileSource, func(key, value string) error {
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
		return fmt.Errorf("cannot add key %s, another key by that name already exists: %v.", keyName, configMap.Data)
	}
	configMap.Data[keyName] = data
	return nil
}
