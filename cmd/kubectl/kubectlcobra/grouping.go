// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

const GroupingLabel = "kustomize.k8s.io/group-id"

// isGroupingObject returns true if the passed object has the
// grouping label.
// TODO(seans3): Check type is ConfigMap.
func isGroupingObject(obj runtime.Object) bool {
	if obj == nil {
		return false
	}
	accessor, err := meta.Accessor(obj)
	if err == nil {
		labels := accessor.GetLabels()
		_, exists := labels[GroupingLabel]
		if exists {
			return true
		}
	}
	return false
}

// findGroupingObject returns the "Grouping" object (ConfigMap with
// grouping label) if it exists, and a boolean describing if it was found.
func findGroupingObject(infos []*resource.Info) (*resource.Info, bool) {
	for _, info := range infos {
		if info != nil && isGroupingObject(info.Object) {
			return info, true
		}
	}
	return nil, false
}

// sortGroupingObject reorders the infos slice to place the grouping
// object in the first position. Returns true if grouping object found,
// false otherwise.
func sortGroupingObject(infos []*resource.Info) bool {
	for i, info := range infos {
		if info != nil && isGroupingObject(info.Object) {
			// If the grouping object is not already in the first position,
			// swap the grouping object with the first object.
			if i > 0 {
				infos[0], infos[i] = infos[i], infos[0]
			}
			return true
		}
	}
	return false
}
