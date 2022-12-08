// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"regexp"
)

// FixKustomizationPreUnmarshalling modifies the raw data
// before marshalling - e.g. changes old field names to
// new field names.
func FixKustomizationPreUnmarshalling(data []byte) ([]byte, error) {
	deprecatedFieldsMap := map[string]string{
		"imageTags:": "images:",
	}
	for oldname, newname := range deprecatedFieldsMap {
		pattern := regexp.MustCompile(oldname)
		data = pattern.ReplaceAll(data, []byte(newname))
	}
	return data, nil
}
