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

package target

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"

	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/types"
)

func loadKustFile(ldr ifc.Loader, withDefaultNames bool) ([]byte, error) {
	if withDefaultNames {
		for _, kf := range []string{
			constants.KustomizationFileName,
			constants.SecondaryKustomizationFileName} {
			content, err := ldr.Load(kf)
			if err == nil {
				return content, nil
			}
			if !strings.Contains(err.Error(), "no such file or directory") {
				return nil, err
			}
		}
		return nil, fmt.Errorf("no kustomization.yaml file under %s", ldr.Root())
	}
	var kustFiles []string
	files, err := ldr.Files()
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if IsKustomizationFile(ldr, f) {
			kustFiles = append(kustFiles, f)
		}
	}

	switch len(kustFiles) {
	case 0:
		return nil, fmt.Errorf("no kustomization.yaml file under %s", ldr.Root())
	case 1:
		content, err := ldr.Load(kustFiles[0])
		if err != nil {
			return nil, err
		}
		return content, nil
	default:
		return nil, fmt.Errorf("Found multiple files for Kustomization under %s", ldr.Root())
	}
}

// IsKustomizationFile checks if a file
// is a kustomization file by verifying GVK in the file
func IsKustomizationFile(ldr ifc.Loader, filepath string) bool {
	content, err := ldr.Load(filepath)
	if err != nil {
		return false
	}

	var it map[string]interface{}
	err = yaml.Unmarshal(content, &it)
	if err != nil {
		return false
	}
	if val, ok := it["kind"]; ok {
		if kind, ok := val.(string); ok {
			if kind == types.KustomizationKind {
				return true
			}
		}
	}
	return false
}
