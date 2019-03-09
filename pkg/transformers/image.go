/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/pkg/image"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

// imageTransformer replace image names and tags
type imageTransformer struct {
	images     []image.Image
	fieldSpecs []config.FieldSpec
}

var _ Transformer = &imageTransformer{}

// NewImageTransformer constructs an imageTransformer.
func NewImageTransformer(slice []image.Image, fs []config.FieldSpec) (Transformer, error) {
	return &imageTransformer{slice, fs}, nil
}

// Transform finds the matching images and replaces name, tag and/or digest
func (pt *imageTransformer) Transform(m resmap.ResMap) error {
	if len(pt.images) == 0 {
		return nil
	}
	for id := range m {
		objMap := m[id].Map()
		for _, path := range pt.fieldSpecs {
			if !id.Gvk().IsSelected(&path.Gvk) {
				continue
			}
			err := mutateField(objMap, path.PathSlice(), false, pt.updateContainers)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func (pt *imageTransformer) updateContainers(in interface{}) (interface{}, error) {
	containers, ok := in.([]interface{})
	if !ok {
		return nil, fmt.Errorf("containers path is not of type []interface{} but %T", in)
	}
	for i := range containers {
		container := containers[i].(map[string]interface{})
		containerImage, found := container["image"]
		if !found {
			continue
		}

		imageName := containerImage.(string)
		for _, img := range pt.images {
			if !isImageMatched(imageName, img.Name) {
				continue
			}
			name, tag := split(imageName)
			if img.NewName != "" {
				name = img.NewName
			}
			if img.NewTag != "" {
				tag = ":" + img.NewTag
			}
			if img.Digest != "" {
				tag = "@" + img.Digest
			}
			container["image"] = name + tag
			break
		}
	}
	return containers, nil
}

func isImageMatched(s, t string) bool {
	// Tag values are limited to [a-zA-Z0-9_.-].
	pattern, _ := regexp.Compile("^" + t + "(:[a-zA-Z0-9_.-]*)?$")
	return pattern.MatchString(s)
}

// split separates and returns the name and tag parts
// from the image string using either colon `:` or at `@` separators.
// Note that the returned tag keeps its separator.
func split(imageName string) (name string, tag string) {
	// check if image name contains a domain
	// if domain is present, ignore domain and check for `:`
	ic := -1
	if slashIndex := strings.Index(imageName, "/"); slashIndex < 0 {
		ic = strings.LastIndex(imageName, ":")
	} else {
		lastIc := strings.LastIndex(imageName[slashIndex:], ":")
		// set ic only if `:` is present
		if lastIc > 0 {
			ic = slashIndex + lastIc
		}
	}
	ia := strings.LastIndex(imageName, "@")
	if ic < 0 && ia < 0 {
		return imageName, ""
	}

	i := ic
	if ic < 0 {
		i = ia
	}

	name = imageName[:i]
	tag = imageName[i:]
	return
}
