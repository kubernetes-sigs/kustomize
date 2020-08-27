// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// SubDirsWithFile takes the root directory path and returns all the paths of
// sub-directories which contain file with input fileName including itself
func SubDirsWithFile(root, fileName string) ([]string, error) {
	var res []string
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, fileName) {
				if root == "." {
					path = root + "/" + path
				}
				res = append(res, path)
			}
			return nil
		})
	if err != nil {
		return res, err
	}
	return res, nil
}
