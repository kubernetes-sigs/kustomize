// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
)

func TestRun(t *testing.T) {
	prc := frameworktestutil.CommandResultsChecker{
		Command: buildCmd,
	}
	prc.Assert(t)
}
