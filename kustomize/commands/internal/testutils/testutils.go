// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package testutils_test

import (
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
)

const (
	// kustomizationContent is used in tests.
	kustomizationContent = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: some-prefix
nameSuffix: some-suffix
# Labels to add to all objects and selectors.
# These labels would also be used to form the selector for apply --prune
# Named differently than “labels” to avoid confusion with metadata for this object
commonLabels:
  app: helloworld
commonAnnotations:
  note: This is an example annotation
resources: []
#- service.yaml
#- ../some-dir/
# There could also be configmaps in Base, which would make these overlays
configMapGenerator: []
# There could be secrets in Base, if just using a fork/rebase workflow
secretGenerator: []
`
)

// WriteTestKustomization writes a standard test file.
func WriteTestKustomization(fSys filesys.FileSystem) {
	WriteTestKustomizationWith(fSys, []byte(kustomizationContent))
}

// WriteTestKustomizationWith writes content to a well known file name.
func WriteTestKustomizationWith(fSys filesys.FileSystem, bytes []byte) {
	fSys.WriteFile(konfig.DefaultKustomizationFileName(), bytes)
}

// ReadTestKustomization reads content from a well known file name.
func ReadTestKustomization(fSys filesys.FileSystem) ([]byte, error) {
	return fSys.ReadFile(konfig.DefaultKustomizationFileName())
}
