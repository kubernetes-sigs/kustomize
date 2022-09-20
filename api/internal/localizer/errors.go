// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"sigs.k8s.io/kustomize/kyaml/errors"
)

var (
	ErrInvalidRoot       = errors.Errorf("invalid root reference")
	ErrLocalizeDirExists = errors.Errorf("'%s' localize directory already exists", LocalizeDir)
	ErrNoRef             = errors.Errorf("localize remote root missing ref query string parameter")
)
