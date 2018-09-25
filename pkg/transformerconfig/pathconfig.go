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

package transformerconfig

import (
	"strings"

	"sigs.k8s.io/kustomize/pkg/gvk"
)

// PathConfig contains the configuration of a field, including the gvk it ties to,
// path to the field, etc.
type PathConfig struct {
	gvk.Gvk            `json:",inline,omitempty" yaml:",inline,omitempty"`
	Path               string `json:"path,omitempty" yaml:"path,omitempty"`
	CreateIfNotPresent bool   `json:"create,omitempty" yaml:"create,omitempty"`
}

// PathSlice converts the path string to a slice of strings, separated by "/"
func (p PathConfig) PathSlice() []string {
	return strings.Split(p.Path, "/")
}
