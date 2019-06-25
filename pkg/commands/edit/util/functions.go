// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"log"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

func GlobPatterns(fsys fs.FileSystem, patterns []string) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		files, err := fsys.Glob(pattern)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			log.Printf("%s has no match", pattern)
			continue
		}
		result = append(result, files...)
	}
	return result, nil
}
