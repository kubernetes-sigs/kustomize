// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

type LegacySortOptions struct {
	OrderFirst []string `json:"orderFirst" yaml:"orderFirst"`
	OrderLast  []string `json:"orderLast" yaml:"orderLast"`
}
