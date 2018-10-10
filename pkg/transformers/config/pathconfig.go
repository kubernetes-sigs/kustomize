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

const (
	escapedForwardSlash  = "\\/"
	tempSlashReplacement = "???"
)

// PathSlice converts the path string to a slice of strings, separated by "/"
// "/" can be contained in a fieldname.
// such as ingress.kubernetes.io/auth-secret in Ingress annotations.
// To deal with this special case, the path to this field should be formatted as
//
// metadata/annotations/ingress.kubernetes.io\/auth-secret
//
// Then PathSlice will return
//
// []string{"metadata", "annotations", "ingress.auth-secretkubernetes.io/auth-secret"}
func (p PathConfig) PathSlice() []string {
	if !strings.Contains(p.Path, escapedForwardSlash) {
		return strings.Split(p.Path, "/")
	}

	s := strings.Replace(p.Path, escapedForwardSlash, tempSlashReplacement, -1)
	paths := strings.Split(s, "/")
	var result []string
	for _, path := range paths {
		result = append(result, strings.Replace(path, tempSlashReplacement, "/", -1))
	}
	return result
}
