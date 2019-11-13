// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package docs

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/kyaml/docs/gen"
)

var Docs = getDocs()

func getDocs() []*cobra.Command {
	return []*cobra.Command{{
		Use:   "docs-merge",
		Short: "Display docs for 2-way merge spec and semantics.",
		Long:  gen.Merge2,
	}, {
		Use:   "docs-merge3",
		Short: "Display docs for 3-way merge spec and semantics.",
		Long:  gen.Merge3,
	}, {
		Use:   "docs-config-fns",
		Short: "Display docs for the Configuration Function spec and semantics.",
		Long:  gen.ConfigFn,
	}, {
		Use:   "docs-config-io",
		Short: "Display docs for configuration input / output spec and semantics",
		Long:  gen.ConfigIo,
	},
	}
}
