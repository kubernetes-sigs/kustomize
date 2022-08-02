// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func cleanDst(newDir string, fSys filesys.FileSystem) {
	if err := fSys.RemoveAll(newDir); err != nil {
		log.Printf("%s", errors.WrapPrefixf(err, "unable to clean localize destination").Error())
	}
}
