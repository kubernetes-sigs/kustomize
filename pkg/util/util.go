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

package util

import (
	"bytes"
	"sort"

	"github.com/ghodss/yaml"

	"k8s.io/kubectl/pkg/kustomize/resource"
	"k8s.io/kubectl/pkg/kustomize/types"
)

// Encode encodes the map `in` and output the encoded objects separated by `---`.
func Encode(in resource.ResourceCollection) ([]byte, error) {
	gvknList := []types.GroupVersionKindName{}
	for gvkn := range in {
		gvknList = append(gvknList, gvkn)
	}
	sort.Sort(types.ByGVKN(gvknList))

	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, gvkn := range gvknList {
		obj := in[gvkn].Data
		out, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}
		if !firstObj {
			_, err = buf.WriteString("---\n")
			if err != nil {
				return nil, err
			}
		}
		_, err = buf.Write(out)
		if err != nil {
			return nil, err
		}
		firstObj = false
	}
	return buf.Bytes(), nil
}

// WriteToDir write each object in ResourceCollection to a file named with GroupVersionKindName.
func WriteToDir(in resource.ResourceCollection, dirName string, printer Printer) (*Directory, error) {
	dir, err := CreateDirectory(dirName)
	if err != nil {
		return &Directory{}, err
	}

	for gvkn, obj := range in {
		f, err := dir.NewFile(gvkn.String())
		if err != nil {
			return &Directory{}, err
		}
		defer f.Close()
		err = printer.Print(obj.Data, f)
		if err != nil {
			return &Directory{}, err
		}
	}
	return dir, nil
}
