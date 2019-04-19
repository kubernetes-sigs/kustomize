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

package loader

import (
	"fmt"

	"sigs.k8s.io/kustomize/pkg/fs"
)

type LoadRestrictorFunc func(
	fs.FileSystem, fs.ConfirmedDir, string) (string, error)

func RestrictionRootOnly(
	fSys fs.FileSystem, root fs.ConfirmedDir, path string) (string, error) {
	d, f, err := fSys.CleanedAbs(path)
	if err != nil {
		return "", err
	}
	if f == "" {
		return "", fmt.Errorf("'%s' must be a file", path)
	}
	if !d.HasPrefix(root) {
		return "", fmt.Errorf(
			"security; file '%s' is not in or below '%s'",
			path, root)
	}
	return d.Join(f), nil
}

func RestrictionNone(
	_ fs.FileSystem, _ fs.ConfirmedDir, path string) (string, error) {
	return path, nil
}
