// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// A strategic merge patch directive.
// See https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md
//
//go:generate stringer -type=smpDirective -linecomment
type smpDirective int

const (
	smpUnknown smpDirective = iota // unknown
	smpReplace                     // replace
	smpDelete                      // delete
	smpMerge                       // merge
)

const strategicMergePatchDirectiveKey = "$patch"

// Examine patch for a strategic merge patch directive.
// If found, return it, and remove the directive from the patch.
func determineSmpDirective(patch *yaml.RNode) (smpDirective, error) {
	if patch == nil {
		return smpMerge, nil
	}
	switch patch.YNode().Kind {
	case yaml.SequenceNode:
		return determineSequenceNodePatchStrategy(patch)
	case yaml.MappingNode:
		return determineMappingNodePatchStrategy(patch)
	default:
		return smpUnknown, fmt.Errorf(
			"no implemented strategic merge patch strategy for '%s' ('%s')",
			patch.YNode().ShortTag(), patch.MustString())
	}
}

// TODO: what should this do?
func determineSequenceNodePatchStrategy(_ *yaml.RNode) (smpDirective, error) {
	return smpMerge, nil
}

func determineMappingNodePatchStrategy(patch *yaml.RNode) (smpDirective, error) {
	node, err := patch.Pipe(yaml.Get(strategicMergePatchDirectiveKey))
	if err != nil || node == nil || node.YNode() == nil {
		return smpMerge, nil
	}
	v := node.YNode().Value
	if v == smpDelete.String() {
		return smpDelete, elidePatchDirective(patch)
	}
	if v == smpReplace.String() {
		return smpReplace, elidePatchDirective(patch)
	}
	if v == smpMerge.String() {
		return smpMerge, elidePatchDirective(patch)
	}
	return smpUnknown, fmt.Errorf(
		"unknown patch strategy '%s'", v)
}

func elidePatchDirective(patch *yaml.RNode) error {
	return patch.PipeE(yaml.Clear(strategicMergePatchDirectiveKey))
}
