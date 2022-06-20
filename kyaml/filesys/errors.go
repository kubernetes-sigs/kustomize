// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filesys

import "sigs.k8s.io/kustomize/kyaml/errors"

var ErrNotDir = errors.Errorf("invalid directory")
