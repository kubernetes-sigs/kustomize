// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kustomize/api/krusty"
)

const flagReorderOutputName = "reorder"

func AddFlagReorderOutput(set *pflag.FlagSet) {
	set.StringVar(
		&theFlags.reorderOutput,
		flagReorderOutputName,
		string(krusty.ReorderOptionUnspecified),
		"enable adding to the new 'sortOptions' field in kustomization.yaml instead")
}
