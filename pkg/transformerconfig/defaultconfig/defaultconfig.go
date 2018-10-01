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

// Package defaultconfig provides the default
// transformer configurations
package defaultconfig

import (
	"bytes"
)

// GetDefaultPathConfigs returns the default pathConfigs data
func GetDefaultPathConfigs() []byte {
	configData := [][]byte{
		[]byte(namePrefixPathConfigs),
		[]byte(commonLabelPathConfigs),
		[]byte(commonAnnotationPathConfigs),
		[]byte(namespacePathConfigs),
		[]byte(varReferencePathConfigs),
		[]byte(nameReferencePathConfigs),
	}
	return bytes.Join(configData, []byte("\n"))
}

// GetDefaultPathConfigStrings returns the default pathConfigs in string format
func GetDefaultPathConfigStrings() map[string]string {
	result := make(map[string]string)
	result["nameprefix"] = namePrefixPathConfigs
	result["commonlabels"] = commonLabelPathConfigs
	result["commonannotations"] = commonAnnotationPathConfigs
	result["namespace"] = namespacePathConfigs
	result["varreference"] = varReferencePathConfigs
	result["namereference"] = namespacePathConfigs
	return result
}
