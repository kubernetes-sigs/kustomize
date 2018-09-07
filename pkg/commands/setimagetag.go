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

package commands

import (
	"errors"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
)

type setImageTagOptions struct {
	imageTagMap map[string]types.ImageTag
}

var pattern = regexp.MustCompile("^(.*):([a-zA-Z0-9._-]*)$")

// newCmdSetImageTag sets the new tags for images in the kustomization.
func newCmdSetImageTag(fsys fs.FileSystem) *cobra.Command {
	var o setImageTagOptions

	cmd := &cobra.Command{
		Use:   "imagetag",
		Short: "Sets images and their new tags or digests in the kustomization file",
		Example: `
The command
  set imagetag nginx:1.8.0 my-app:latest alpine@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3
will add 

imageTags:
- name: nginx
  newTag: 1.8.0
- name: my-app
  newTag: latest
- name: alpine
  digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3

to the kustomization file if it doesn't exist,
and overwrite the previous ones if the image tag exists.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunSetImageTags(fsys)
		},
	}
	return cmd
}

// Validate validates setImageTag command.
func (o *setImageTagOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("no image and newTag specified")
	}

	o.imageTagMap = make(map[string]types.ImageTag)

	for _, arg := range args {
		if strings.Contains(arg, "@") {
			img := strings.Split(arg, "@")
			o.imageTagMap[img[0]] = types.ImageTag{
				Name:   img[0],
				Digest: img[1],
			}
			continue
		}

		imagetag := pattern.FindStringSubmatch(arg)
		if len(imagetag) != 3 {
			return errors.New("invalid format of imagetag, must specify it as <image>:<newtag> or <image>@<digest>")
		}
		o.imageTagMap[imagetag[1]] = types.ImageTag{
			Name:   imagetag[1],
			NewTag: imagetag[2],
		}
	}
	return nil
}

// RunSetImageTags runs setImageTags command (does real work).
func (o *setImageTagOptions) RunSetImageTags(fsys fs.FileSystem) error {
	mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
	if err != nil {
		return err
	}
	m, err := mf.read()
	if err != nil {
		return err
	}

	for _, it := range m.ImageTags {
		if _, ok := o.imageTagMap[it.Name]; ok {
			continue
		}

		o.imageTagMap[it.Name] = it
	}

	var imageTags []types.ImageTag
	for _, v := range o.imageTagMap {
		imageTags = append(imageTags, v)
	}
	sort.Slice(imageTags, func(i, j int) bool {
		return imageTags[i].Name < imageTags[j].Name
	})

	m.ImageTags = imageTags

	return mf.write(m)
}
