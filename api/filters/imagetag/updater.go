// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package imagetag

import (
	"sigs.k8s.io/kustomize/api/image"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// imageTagUpdater is an implementation of the kio.Filter interface
// that will update the value of the yaml node based on the provided
// ImageTag if the current value matches the format of an image reference.
type imageTagUpdater struct {
	Kind     string      `yaml:"kind,omitempty"`
	ImageTag types.Image `yaml:"imageTag,omitempty"`
}

func (u imageTagUpdater) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	if err := yaml.ErrorIfInvalid(rn, yaml.ScalarNode); err != nil {
		return nil, err
	}

	value := rn.YNode().Value

	if !image.IsImageMatched(value, u.ImageTag.Name) {
		return rn, nil
	}

	name, tag, digest := image.Split(value)
	if u.ImageTag.NewName != "" {
		name = u.ImageTag.NewName
	}
	// not overriding tag/digest component, keep original
	if u.ImageTag.NewTag == "" && u.ImageTag.Digest == "" {
		if tag != "" {
			tag = ":" + tag
		}
		if digest != "" {
			tag += "@" + digest
		}
	}
	// overriding tag or digest will replace both original tag and digest values
	if u.ImageTag.NewTag != "" && u.ImageTag.Digest != "" {
		tag = ":" + u.ImageTag.NewTag + "@" + u.ImageTag.Digest
	} else if u.ImageTag.NewTag != "" {
		tag = ":" + u.ImageTag.NewTag
	} else if u.ImageTag.Digest != "" {
		tag = "@" + u.ImageTag.Digest
	}

	return rn.Pipe(yaml.FieldSetter{StringValue: name + tag})
}
