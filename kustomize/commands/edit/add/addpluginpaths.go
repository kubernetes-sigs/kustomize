// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"log"
	"slices"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func addPluginPaths(
	fSys filesys.FileSystem,
	paths []string,
	kind string,
	selectPaths func(*types.Kustomization) *[]string,
) error {
	if len(paths) == 0 {
		return nil
	}
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err //nolint:wrapcheck // Preserve the edit command's existing error text.
	}
	m, err := mf.Read()
	if err != nil {
		return err //nolint:wrapcheck // Preserve the edit command's existing error text.
	}
	entries := selectPaths(m)
	for _, path := range paths {
		if slices.Contains(*entries, path) {
			log.Printf("%s %s already in kustomization file", kind, path)
			continue
		}
		*entries = append(*entries, path)
	}
	return mf.Write(m) //nolint:wrapcheck // Preserve the edit command's existing error text.
}
