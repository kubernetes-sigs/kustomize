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

	"github.com/spf13/cobra"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
)

type setImageTagOptions struct {
	imageTagMap map[string]string
}

var pattern = regexp.MustCompile("^(.*):([a-zA-Z0-9._-]*)$")

// newCmdSetImageTag sets the new tags for images in the kustomization.
func newCmdSetImageTag(fsys fs.FileSystem) *cobra.Command {
	var o setImageTagOptions

	cmd := &cobra.Command{
		Use:   "imagetag",
		Short: "Sets images and their new tags in the kustomization file",
		Example: `
The command
  set imagetag nginx:1.8.0 my-app:latest
will add 

imageTags:
- name: nginx
  newTag: 1.8.0
- name: my-app
  newTag: latest

to the kustomization file if it doesn't exist,
and overwrite the previous newTag if the image name exists.
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
	o.imageTagMap = make(map[string]string)
	for _, arg := range args {
		imagetag := pattern.FindStringSubmatch(arg)
		if len(imagetag) != 3 {
			return errors.New("invalid format of imagetag, must specify it as <image>:<newtag>")
		}
		o.imageTagMap[imagetag[1]] = imagetag[2]
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
	imageTagMap := map[string]string{}
	for _, it := range m.ImageTags {
		imageTagMap[it.Name] = it.NewTag
	}
	for key, value := range o.imageTagMap {
		imageTagMap[key] = value
	}
	var imageTags []types.ImageTag
	for key, value := range imageTagMap {
		imageTags = append(imageTags, types.ImageTag{Name: key, NewTag: value})
	}
	sort.Slice(imageTags, func(i, j int) bool {
		return imageTags[i].Name < imageTags[j].Name
	})

	m.ImageTags = imageTags

	return mf.write(m)
}
