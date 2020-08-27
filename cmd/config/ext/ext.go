// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package ext

import (
	"path/filepath"
)

// GetOpenAPIFile returns the path to the file containing supplementary OpenAPI definitions.
// Maybe be overridden to configure which file to read OpenAPI definitions from.
var GetOpenAPIFile = func(args []string) (string, error) {
	return filepath.Join(args[0], "Krmfile"), nil
}

// OpenAPIFileName returns the name of the file with openAPI definitions
// uses OpenAPIFile function to derive it
func OpenAPIFileName() (string, error) {
	openAPIFileName, err := GetOpenAPIFile([]string{"."})
	if err != nil {
		return "", err
	}
	return openAPIFileName, nil
}
