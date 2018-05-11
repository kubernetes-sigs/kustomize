/*
Copyright 2015 The Kubernetes Authors.

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

// HandleFromLiteralSources adds the specified literal source information into the provided secret
func HandleFromLiteralSources(secret *v1.Secret, literalSources []string) error {
	for _, literalSource := range literalSources {
		keyName, value, err := ParseLiteralSource(literalSource)
		if err != nil {
			return err
		}
		if err = addKeyFromLiteralToSecret(secret, keyName, []byte(value)); err != nil {
			return err
		}
	}
	return nil
}

// HandleFromFileSources adds the specified file source information into the provided secret
func HandleFromFileSources(secret *v1.Secret, fileSources []string) error {
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
					if err = addKeyFromFileToSecret(secret, keyName, itemPath); err != nil {
						return err
					}
				}
			}
		} else {
			if err := addKeyFromFileToSecret(secret, keyName, filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

// HandleFromEnvFileSource adds the specified env file source information
// into the provided secret
func HandleFromEnvFileSource(secret *v1.Secret, envFileSource string) error {
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
		return fmt.Errorf("env secret file cannot be a directory")
	}

	return addFromEnvFile(envFileSource, func(key, value string) error {
		return addKeyFromLiteralToSecret(secret, key, []byte(value))
	})
}

func addKeyFromFileToSecret(secret *v1.Secret, keyName, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return addKeyFromLiteralToSecret(secret, keyName, data)
}

func addKeyFromLiteralToSecret(secret *v1.Secret, keyName string, data []byte) error {
	if errs := validation.IsConfigMapKey(keyName); len(errs) != 0 {
		return fmt.Errorf("%q is not a valid key name for a Secret: %s", keyName, strings.Join(errs, ";"))
	}

	if _, entryExists := secret.Data[keyName]; entryExists {
		return fmt.Errorf("cannot add key %s, another key by that name already exists: %v.", keyName, secret.Data)
	}
	secret.Data[keyName] = data
	return nil
}
