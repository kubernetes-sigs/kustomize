// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// NewCmdAdd returns an instance of 'add' subcommand.
func NewCmdAdd(
	fSys filesys.FileSystem,
	ldr ifc.KvLoader,
	rf *resource.Factory) *cobra.Command {
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
	kustomize edit add patch --path {filepath} --group {target group name} --version {target version}

	# Adds a component to the kustomization
	kustomize edit add component <filepath>

	# Adds one or more base directories to the kustomization
	kustomize edit add base <filepath>
	kustomize edit add base <filepath1>,<filepath2>,<filepath3>

	# Adds one or more commonLabels to the kustomization
	kustomize edit add label {labelKey1:labelValue1},{labelKey2:labelValue2}

	# Adds one or more commonAnnotations to the kustomization
	kustomize edit add annotation {annotationKey1:annotationValue1},{annotationKey2:annotationValue2}

	# Adds a transformer configuration to the kustomization
	kustomize edit add transformer <filepath>
`,
		Args: cobra.MinimumNArgs(1),
	}
	c.AddCommand(
		newCmdAddResource(fSys),
		newCmdAddPatch(fSys),
		newCmdAddComponent(fSys),
		newCmdAddSecret(fSys, ldr, rf),
		newCmdAddConfigMap(fSys, ldr, rf),
		newCmdAddBase(fSys),
		newCmdAddLabel(fSys, ldr.Validator().MakeLabelValidator()),
		newCmdAddAnnotation(fSys, ldr.Validator().MakeAnnotationValidator()),
		newCmdAddTransformer(fSys),
	)
	return c
}
