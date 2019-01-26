/*
Copyright 2017 The Kubernetes Authors.

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

package add

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

// NewCmdAdd returns an instance of 'add' subcommand.
func NewCmdAdd(fsys fs.FileSystem, v ifc.Validator, kf ifc.KunstructuredFactory) *cobra.Command {
	c := &cobra.Command{
		Use:   "add",
		Short: "Adds an item to the kustomization file.",
		Long:  "",
		Example: `
	# Adds a secret to the kustomization file
	kustomize edit add secret NAME --from-literal=k=v

	# Adds a configmap to the kustomization file
	kustomize edit add configmap NAME --from-literal=k=v

	# Adds a resource to the kustomization
	kustomize edit add resource <filepath>

	# Adds a patch to the kustomization
	kustomize edit add patch <filepath>

	# Adds one or more base directories to the kustomization
	kustomize edit add base <filepath>
	kustomize edit add base <filepath1>,<filepath2>,<filepath3>

	# Adds one or more commonLabels to the kustomization
	kustomize edit add label {labelKey1:labelValue1},{labelKey2:labelValue2}

	# Adds one or more commonAnnotations to the kustomization
	kustomize edit add annotation {annotationKey1:annotationValue1},{annotationKey2:annotationValue2}
`,
		Args: cobra.MinimumNArgs(1),
	}
	c.AddCommand(
		newCmdAddResource(fsys),
		newCmdAddPatch(fsys),
		newCmdAddSecret(fsys, kf),
		newCmdAddConfigMap(fsys, kf),
		newCmdAddBase(fsys),
		newCmdAddLabel(fsys, v.MakeLabelValidator()),
		newCmdAddAnnotation(fsys, v.MakeAnnotationValidator()),
	)
	return c
}
