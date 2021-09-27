// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package pathutil

import (
	"os"
	"path/filepath"
)

// DirsWithFile takes the root directory path and returns all the paths of
// sub-directories(including itself) which contain file with input fileName
// at top level if recurse is true
func DirsWithFile(root, fileName string, recurse bool) ([]string, error) {
	var res []string
	if !recurse {
		// check if the file with fileName is present in root and return it
		// else return empty list
		_, err := os.Stat(filepath.Join(root, fileName))
		if !os.IsNotExist(err) {
			res = append(res, filepath.Clean(root))
		}
		return res, nil
	}
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Base(path) == fileName {
				res = append(res, filepath.Dir(path))
			}
			return nil
		})
	if err != nil {
		return res, err
	}
	return res, nil
}
