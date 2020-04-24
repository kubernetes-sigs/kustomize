// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
package fsslice

import (
	"strings"

	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// parseGV parses apiVersion field into group and version.
func parseGV(apiVersion string) (group, version string) {
	// parse the group and version from the apiVersion field
	parts := strings.SplitN(apiVersion, "/", 2)
	group = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	// TODO: Special case the original "apiVersion" of what
	//       we now call the "core" (empty) group.
	//if group == "v1" && version == "" {
	//	version = "v1"
	//	group = ""
	//}
	return
}

// GetGVK parses the metadata into a GVK
func GetGVK(meta yaml.ResourceMeta) resid.Gvk {
	group, version := parseGV(meta.APIVersion)
	return resid.Gvk{
		Group:   group,
		Version: version,
		Kind:    meta.Kind,
	}
}
