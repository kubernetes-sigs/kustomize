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

package transformers

import (
	"strings"

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
)

// imageTagTransformer replace image tags
type imageTagTransformer struct {
	imageTags []types.ImageTag
}

var _ Transformer = &imageTagTransformer{}

// NewImageTagTransformer constructs a imageTagTransformer.
func NewImageTagTransformer(slice []types.ImageTag) (Transformer, error) {
	return &imageTagTransformer{slice}, nil
}

// Transform finds the matching images and replace the tag
func (pt *imageTagTransformer) Transform(resources resmap.ResMap) error {
	if len(pt.imageTags) == 0 {
		return nil
	}
	for _, res := range resources {
		err := pt.findAndReplaceTag(res.UnstructuredContent())
		if err != nil {
			return err
		}
	}
	return nil
}

/*
 findAndReplaceTag replaces the image tags inside one object
 It searches the object for container session
 then loops though all images inside containers session, finds matched ones and update the tag name
*/
func (pt *imageTagTransformer) findAndReplaceTag(obj map[string]interface{}) error {
	_, found := obj["containers"]
	if found {
		return pt.updateContainers(obj)
	}
	return pt.findContainers(obj)
}

func (pt *imageTagTransformer) updateContainers(obj map[string]interface{}) error {
	containers := obj["containers"].([]interface{})
	for i := range containers {
		container := containers[i].(map[string]interface{})
		image, found := container["image"]
		if !found {
			continue
		}
		for _, imagetag := range pt.imageTags {
			if isImageMatched(image.(string), imagetag.Name) {
				container["image"] = strings.Join([]string{imagetag.Name, imagetag.NewTag}, ":")
				break
			}
		}
	}
	return nil
}

func (pt *imageTagTransformer) findContainers(obj map[string]interface{}) error {
	for key := range obj {
		switch typedV := obj[key].(type) {
		case map[string]interface{}:
			err := pt.findAndReplaceTag(typedV)
			if err != nil {
				return err
			}
		case []interface{}:
			for i := range typedV {
				item := typedV[i]
				typedItem, ok := item.(map[string]interface{})
				if ok {
					err := pt.findAndReplaceTag(typedItem)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func isImageMatched(s, t string) bool {
	imagetag := strings.Split(s, ":")
	return len(imagetag) >= 1 && imagetag[0] == t
}
