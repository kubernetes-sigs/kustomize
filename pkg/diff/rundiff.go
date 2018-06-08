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

// Package diff runs system `diff` to compare resource collections.
package diff

import (
	"github.com/ghodss/yaml"

	"io"

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
)

// RunDiff runs system diff program to compare two Maps.
func RunDiff(raw, transformed resmap.ResMap, out, errOut io.Writer) error {
	transformedDir, err := writeYamlToNewDir(transformed, "transformed")
	if err != nil {
		return err
	}
	defer transformedDir.delete()

	noopDir, err := writeYamlToNewDir(raw, "noop")
	if err != nil {
		return err
	}
	defer noopDir.delete()

	return newProgram(out, errOut).run(noopDir.name(), transformedDir.name())
}

// writeYamlToNewDir writes each obj in ResMap to a file in a new directory.
// The directory's name will begin with the given prefix.
// Each file is named with GroupVersionKindName.
func writeYamlToNewDir(in resmap.ResMap, prefix string) (*directory, error) {
	dir, err := newDirectory(prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range in {
		f, err := dir.newFile(id.String())
		if err != nil {
			return nil, err
		}
		err = print(obj, f)
		f.Close()
		if err != nil {
			return nil, err
		}
	}
	return dir, nil
}

// Print the object as YAML.
func print(obj interface{}, w io.Writer) error {
	if obj == nil {
		return nil
	}
	data, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
