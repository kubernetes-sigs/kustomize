// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"slices"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func removePathsFromKustomization(
	fSys filesys.FileSystem,
	patterns []string,
	selectPaths func(*types.Kustomization) *[]string,
) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err //nolint:wrapcheck // Preserve the edit command's existing error text.
	}
	m, err := mf.Read()
	if err != nil {
		return err //nolint:wrapcheck // Preserve the edit command's existing error text.
	}
	entries := selectPaths(m)
	matches, err := globPatterns(*entries, patterns)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return nil
	}
	remaining := make([]string, 0, len(*entries))
	for _, path := range *entries {
		if !slices.Contains(matches, path) {
			remaining = append(remaining, path)
		}
	}
	*entries = remaining
	return mf.Write(m) //nolint:wrapcheck // Preserve the edit command's existing error text.
}
