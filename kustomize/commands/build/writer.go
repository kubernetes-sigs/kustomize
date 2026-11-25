// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/yaml"
)

type Writer struct {
	fSys filesys.FileSystem
}

func MakeWriter(fSys filesys.FileSystem) *Writer {
	return &Writer{
		fSys: fSys,
	}
}

func (w Writer) WriteIndividualFiles(dirPath string, m resmap.ResMap) error {
	type generatedNameOutput struct {
		stem string
		res  *resource.Resource
	}
	var generated []generatedNameOutput
	usedNames := make(map[string]struct{})
	multipleNamespaces := len(m.GroupedByCurrentNamespace()) > 1
	for _, res := range m.Resources() {
		name := fileName(res)
		if namespace := res.CurId().EffectiveNamespace(); multipleNamespaces && namespace != resid.TotallyNotANamespace {
			name = strings.ToLower(namespace) + "_" + name
		}
		if res.GetName() == "" {
			generated = append(generated, generatedNameOutput{strings.TrimSuffix(name, ".yaml"), res})
			continue
		}
		usedNames[name] = struct{}{}
		if err := w.write(dirPath, name, res); err != nil {
			return err
		}
	}

	// Reserve numbered filenames for resources whose names will be generated
	// on the server, skipping filenames already used by named resources.
	lastIndex := make(map[string]int)
	for _, output := range generated {
		var name string
		for {
			lastIndex[output.stem]++
			name = fmt.Sprintf("%s_%d.yaml", output.stem, lastIndex[output.stem])
			if _, exists := usedNames[name]; !exists {
				break
			}
		}
		usedNames[name] = struct{}{}
		if err := w.write(dirPath, name, output.res); err != nil {
			return err
		}
	}
	return nil
}

func (w Writer) write(path, fName string, res *resource.Resource) error {
	m, err := res.Map()
	if err != nil {
		return err
	}
	yml, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return w.fSys.WriteFile(filepath.Join(path, fName), yml)
}

func fileName(res *resource.Resource) string {
	return strings.ToLower(res.GetGvk().StringWoEmptyField()) +
		"_" + strings.ToLower(res.GetName()) + ".yaml"
}
