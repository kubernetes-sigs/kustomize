// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// +build tools

// This file exists to declare that package
// plugin explicitly depends on the pluginator
// tool (via go:generate directives)
package plugin

import (
	_ "sigs.k8s.io/kustomize/pluginator"
)
