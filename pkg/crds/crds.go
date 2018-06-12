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

// Package crds read in files for CRD schemas and parse annotations from it
package crds

import (
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
)

// RegisterCRDs parse CRD schemas from paths and update various pathConfigs
func RegisterCRDs(loader loader.Loader, paths []string) error {
	return nil
}
